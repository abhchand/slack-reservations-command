package main

import (
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "time"
)

var unit_standardization_mapping = map[string]string{
    "min":      "minute",
    "mins":     "minute",
    "minute":   "minute",
    "minutes":  "minute",
    "hr":       "hour",
    "hrs":      "hour",
    "hour":     "hour",
    "hours":    "hour",
}

var reservations_dir    = "/tmp"
var reservations_file   = filepath.Join(reservations_dir, "reservations.json")

var subcmd_help_regex       = regexp.MustCompile("\\Ahelp\\z")
var subcmd_status_regex     = regexp.MustCompile("\\Alist\\z")
var subcmd_reserve_regex    = regexp.MustCompile("\\Acreate (.*) (\\d*) (mins?|minutes?|hrs?|hours?)\\z")
var subcmd_extend_regex     = regexp.MustCompile("\\Aextend (.*) (\\d*) (mins?|minutes?|hrs?|hours?)\\z")
var subcmd_cancel_regex     = regexp.MustCompile("\\Acancel (.*)\\z")

func MainHandler(w http.ResponseWriter, r *http.Request) {

    // Parse incoming slack request data
    slack_request, err := parseSlackRequest(r)
    if err != nil { buildInvalidResponse(w); return }

    // Check validity of slack verification token
    if !isValidSlackVerificationToken(slack_request) {
        buildInvalidResponse(w); return
    }

    // Create reservations file if it doesn't exist
    log.Debug("Ensuring file exists...")
    err = ensureReservationsFileExists()
    if err != nil { buildErrorResponse(w); return }

    // Call appropriate command handler
    command := slack_request.FormattedSubcommand()
    var slack_response SlackResponse
    var success bool

    switch {

    case subcmd_help_regex.MatchString(command):
        log.Debug("Handling command: `help`")
        slack_response, success = handleCommandHelp(slack_request)

    case subcmd_status_regex.MatchString(command):
        log.Debug("Handling command: `list`")
        slack_response, success = handleCommandList(slack_request)

    case subcmd_reserve_regex.MatchString(command):
        log.Debug("Handling command: `reserve`")
        slack_response, success = handleCommandReserve(slack_request)

    case subcmd_extend_regex.MatchString(command):
        log.Debug("Handling command: `extend`")
        slack_response, success = handleCommandExtend(slack_request)

    case subcmd_cancel_regex.MatchString(command):
        log.Debug("Handling command: `cancel`")
        slack_response, success = handleCommandCancel(slack_request)

    default:
        buildErrorResponse(w)
        return
    }

    if !success { buildErrorResponse(w); return }

    buildResponse(slack_response, w)

}

func isValidSlackVerificationToken(s SlackRequest) bool {

    token := os.Getenv("SLACK_VERIFICATION_TOKEN")
    valid := (token == s.Token)

    if !valid { log.Errorf("Invalid Slack token %v", token) }

    return os.Getenv("SLACK_VERIFICATION_TOKEN") == s.Token

}

func parseSlackRequest(r *http.Request) (SlackRequest, error) {

    var err error
    var slack_request SlackRequest

    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        log.Error("Could not ready request body")
        return slack_request, err
    }

    err = r.Body.Close();
    if err != nil {
        log.Error("Could not close body")
        return slack_request, err
    }

    err = json.Unmarshal(body, &slack_request)
    if err != nil {
        log.Errorf("Could not unmarshal body: %v", body)
        return slack_request, err
    }

    return slack_request, nil

}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/help.json \
     http://localhost:8080/slack/commands/reservations

*/
func handleCommandHelp(slack_request SlackRequest) (SlackResponse, bool) {

    resources := ListOfResources()
    example_resource := resources[0]

    help_text := `
    A basic reservations system for shared resources

*list* - List all resources and any reservations
` + "`/reservations list`" + `

*create* - Create a new reservation
` + "`/reservations create (resource) (duration)`" + `
` + fmt.Sprintf("`/reservations create %v 3 hours`", example_resource) + `

*extend* - Extend an existing reservation
` + "`/reservations extend (resource) (duration)`" + `
` + fmt.Sprintf("`/reservations extend %v 20 mins`", example_resource) + `

*create* - Cancel an existing reservation
` + "`/reservations cancel (resource)`" + `
` + fmt.Sprintf("`/reservations cancel %v`", example_resource) + `

Duration units can be singular or plural form of:
    mins, minutes, hrs, hours
`

    return SlackResponse{Text: help_text}, true
}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/list.json \
     http://localhost:8080/slack/commands/reservations

*/
func handleCommandList(slack_request SlackRequest) (SlackResponse, bool) {

    response := SlackResponse{}

    // Find all reservations
    reservations, err := NewReservations()
    if err != nil {
        log.Error(err)
        return response, false
    }

    response_text := "Reservations\n\n"

    for _, resource := range ListOfResources() {

        reservation := reservations[resource]
        if (reservation != Reservation{}) && reservation.IsActive() {
            response_text += fmt.Sprintf(
                "%v (reserved by @%v, expires in %v)\n",
                resource,
                reservation.User,
                reservation.RemainingTimeToString())
        } else {
            response_text += fmt.Sprintf(
                "%v (free)\n",
                resource)
        }
    }

    response.Text = response_text
    return response, true

}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/reserve.json \
     http://localhost:8080/slack/commands/reservations

*/
func handleCommandReserve(slack_request SlackRequest) (SlackResponse, bool) {

    command  := slack_request.FormattedSubcommand()
    response := SlackResponse{}

    // Extract data from command
    matches     := subcmd_reserve_regex.FindStringSubmatch(command)
    resource    := matches[1]
    time_value  := matches[2]
    unit        := matches[3]

    // If an active reservation already exists against this resource, don't
    // allow a new reservation
    reservations, err := NewReservations()
    if err != nil {
        log.Error(err)
        return response, false
    }

    reservation := reservations.FindByResource(resource)
    if reservation.IsPresent() && reservation.IsActive(){
        if slack_request.UserName == reservation.User {
            response.Text = fmt.Sprintf(
                "You've already reserved resource *%v* for the next *%v*",
                resource,
                reservation.RemainingTimeToString())
        } else {
            response.Text = fmt.Sprintf(
                "@%v has has already reserved resource *%v* for the next *%v*",
                reservation.User,
                resource,
                reservation.RemainingTimeToString())
        }

        return response, true
    }

    // Transform value and units into formats we can work with
    time_value_int, err := strconv.Atoi(time_value)
    if err != nil {
        log.Error(err)
        return response, false
    }
    unit = unit_standardization_mapping[unit]

    // Calculate a new endAt
    hours := 0
    minutes := 0
    if unit == "hour" { hours = time_value_int }
    if unit == "minute" { minutes = time_value_int }

    endAt := time.Now().Add(
        time.Hour * time.Duration(hours) + time.Minute * time.Duration(minutes),
    )

    // Create new reservation
    reservation = Reservation{User: slack_request.UserName, EndAt: endAt}

    // Update and save to file
    reservations.Upsert(resource, reservation)
    reservations.WriteToFile()

    // Construct a response for the user
    response.Text = fmt.Sprintf(
        "You have reserved resource *%v* for the next *%v*",
        resource,
        reservation.RemainingTimeToString())

    return response, true
}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/extend.json \
     http://localhost:8080/slack/commands/reservations

*/
func handleCommandExtend(slack_request SlackRequest) (SlackResponse, bool) {

    command  := slack_request.FormattedSubcommand()
    response := SlackResponse{}

    // Extract data from command
    matches     := subcmd_extend_regex.FindStringSubmatch(command)
    resource    := matches[1]
    time_value  := matches[2]
    unit        := matches[3]

    // Ensure an active reservation exists for this reousrce and user.
    reservations, err := NewReservations()
    if err != nil {
        log.Error(err)
        return response, false
    }

    reservation := reservations.FindByResource(resource)
    if !reservation.IsPresent() ||
            !reservation.IsActive() ||
            slack_request.UserName != reservation.User {
        response.Text = fmt.Sprintf(
                "You do not have any reservation to extend on resource *%v*",
                resource)

        return response, true
    }

    // Transform value and units into formats we can work with
    time_value_int, err := strconv.Atoi(time_value)
    if err != nil {
        log.Error(err)
        return response, false
    }
    unit = unit_standardization_mapping[unit]

    // Calculate a new endAt
    hours := 0
    minutes := 0
    if unit == "hour" { hours = time_value_int }
    if unit == "minute" { minutes = time_value_int }

    endAt := reservation.EndAt.Add(
        time.Hour * time.Duration(hours) + time.Minute * time.Duration(minutes),
    )

    // Update reservation
    reservation.EndAt = endAt

    // Update and save to file
    reservations.Upsert(resource, reservation)
    reservations.WriteToFile()

    // Construct a response for the user
    response.Text = fmt.Sprintf(
        "You have extended your reservation on resource *%v*. It now expires" +
        " in *%v*",
        resource,
        reservation.RemainingTimeToString())

    return response, true

}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/cancel.json \
     http://localhost:8080/slack/commands/reservations

*/
func handleCommandCancel(slack_request SlackRequest) (SlackResponse, bool) {

    command  := slack_request.FormattedSubcommand()
    response := SlackResponse{}

    // Extract data from command
    matches     := subcmd_cancel_regex.FindStringSubmatch(command)
    resource    := matches[1]

    // Ensure an active reservation exists for this reousrce and user.
    reservations, err := NewReservations()
    if err != nil {
        log.Error(err)
        return response, false
    }

    reservation := reservations.FindByResource(resource)
    if !reservation.IsPresent() ||
            !reservation.IsActive() ||
            slack_request.UserName != reservation.User {
        response.Text = fmt.Sprintf(
                "You do not have any reservation to cancel on resource *%v*",
                resource)

        return response, true
    }

    // Cancel reservation and save to file
    reservations.Delete(resource)
    reservations.WriteToFile()

    // Construct a response for the user
    response.Text = fmt.Sprintf(
        "Your reservation on resource *%v* has been cancelled",
        resource)

    return response, true

}

func ensureReservationsFileExists() error {

    var err error

    // Create directory if it does not exist
    _, err = os.Stat(reservations_dir)
    if err != nil && os.IsNotExist(err) {
        err = os.MkdirAll(reservations_dir, 0775)

        if err != nil {
         log.Debug("Error creating directory %v", reservations_dir)
         return err
        }
    }

    // Create file if it does not exist
    _, err = os.Stat(reservations_file)
    if err != nil && os.IsNotExist(err) {
        err = ioutil.WriteFile(reservations_file, []byte("{}"), 0755)
        if err != nil { return err }
    }

    return nil

}


func buildInvalidResponse(w http.ResponseWriter) {

    response := SlackResponse{
        Text: "Yikes, that looks like an invalid request!",
    }
    buildResponse(response, w)

}

func buildErrorResponse(w http.ResponseWriter) {

    response := SlackResponse{
        Text: "Uh oh, there was a problem handling your request",
    }
    buildResponse(response, w)

}

func buildResponse(slack_response SlackResponse, w http.ResponseWriter) {

    // Hard-code response_type as ephemeral for all responses
    slack_response.ResponseType = "ephemeral"

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)

    err := json.NewEncoder(w).Encode(slack_response)
    if err != nil { panic(err) }

}
