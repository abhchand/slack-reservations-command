package main

import (
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
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

    if !valid { log.Errorf("Invalid Slack token %v", s.Token) }

    return os.Getenv("SLACK_VERIFICATION_TOKEN") == s.Token

}

func parseSlackRequest(r *http.Request) (SlackRequest, error) {

    var err error
    var slack_request SlackRequest

    // Read the body
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576 /*1MB*/))
    if err != nil {
        log.Error("Could not ready request body")
        return slack_request, err
    }

    err = r.Body.Close();
    if err != nil {
        log.Error("Could not close body")
        return slack_request, err
    }

    request_str := string(body)
    log.Debugf("Received slack request: \"%v\"", request_str)

    // URL-Decode the request body
    request_str, err = url.PathUnescape(request_str)
    if err != nil { log.Error("Could not unescape request body") }

    // Parse the query parameters
    qp, err := url.ParseQuery(request_str)
    if err != nil { log.Error("Could not parse query params") }

    // Populate the struct
    slack_request.Token           = qp.Get("token")
    slack_request.TeamId          = qp.Get("team_id")
    slack_request.TeamDomain      = qp.Get("team_domain")
    slack_request.EnterpriseId    = qp.Get("enterprise_id")
    slack_request.EnterpriseName  = qp.Get("enterprise_name")
    slack_request.ChannelId       = qp.Get("channel_id")
    slack_request.ChannelName     = qp.Get("channel_name")
    slack_request.UserId          = qp.Get("user_id")
    slack_request.UserName        = qp.Get("user_name")
    slack_request.Command         = qp.Get("command")
    slack_request.Text            = qp.Get("text")
    slack_request.ResponseUrl     = qp.Get("response_url")
    slack_request.TriggerId       = qp.Get("trigger_id")

    return slack_request, nil

}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/help \
     http://localhost:8080/slack/commands/reservations

*/
func handleCommandHelp(slack_request SlackRequest) (SlackResponse, bool) {

    resources := ListOfResources()
    example_resource := resources[0]

    help_text := `
    _*Reservations Bot*_
    A basic reservations system for shared resources

*list* - List all resources and any reservations
` + "`/reservations list`" + `

*create* - Create a new reservation
` + "`/reservations create (resource) (duration - mins/minutes/hrs/hours)`" + `
` + fmt.Sprintf("`/reservations create %v 3 hours`", example_resource) + `

*extend* - Extend an existing reservation
` + "`/reservations extend (resource) (duration - mins/minutes/hrs/hours)`" + `
` + fmt.Sprintf("`/reservations extend %v 20 mins`", example_resource) + `

*cancel* - Cancel an existing reservation
` + "`/reservations cancel (resource)`" + `
` + fmt.Sprintf("`/reservations cancel %v`", example_resource) + `


`

    return SlackResponse{Text: help_text}, true
}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/list \
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

    response_text := "\n_*Reservations*_\n\n"

    for _, resource := range ListOfResources() {

        reservation := reservations[resource]
        if (reservation != Reservation{}) && reservation.IsActive() {
            response_text += fmt.Sprintf(
                "→  %v (reserved by %v, expires in %v)\n",
                resource,
                reservation.User,
                reservation.RemainingTimeToString())
        } else {
            response_text += fmt.Sprintf(
                "→  %v (free)\n",
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
     -d @example/create \
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
                "You've already reserved \"*%v*\" for the next *%v*",
                resource,
                reservation.RemainingTimeToString())
        } else {
            response.Text = fmt.Sprintf(
                "%v has reserved \"*%v*\" for the next *%v*",
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

    // Update
    err = reservations.Upsert(resource, reservation)
    if err != nil {
        if isInvalidResourceError(err) {
            response.Text = unknownResourceText(resource)
            return response, true
        } else {
            log.Error(err)
            return response, false
        }
    }

    // Save to file
    err = reservations.WriteToFile()
    if err != nil { log.Error(err); return response, false }

    // Construct a response for the user
    response.Text = fmt.Sprintf(
        "You've successfully reserved \"*%v*\" for the next *%v*",
        resource,
        reservation.RemainingTimeToString())

    return response, true
}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/extend \
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

    // Check that resource is valid
    // Upsert() below checks for this, but we want to do it sooner so we
    // can return the appropriate message to the user
    if !IsValidResource(resource) {
        response.Text = unknownResourceText(resource)
        return response, true
    }

    // Find all reservations
    reservations, err := NewReservations()
    if err != nil {
        log.Error(err)
        return response, false
    }

    // Ensure an active reservation exists for this reousrce and user.
    reservation := reservations.FindByResource(resource)
    if !reservation.IsPresent() ||
            !reservation.IsActive() ||
            slack_request.UserName != reservation.User {
        response.Text = fmt.Sprintf(
                "You don't have any reservation on \"*%v*\" to extend\n\n" +
                    "Type `/reservations list` to list current reservations",
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

    // Update
    // No need to check explicitly for `isInvalidResourceError()` since
    // that's already done manually above
    err = reservations.Upsert(resource, reservation)
    if err != nil {
        log.Error(err)
        return response, false
    }

    // Save to file
    err = reservations.WriteToFile()
    if err != nil { log.Error(err); return response, false }

    // Construct a response for the user
    response.Text = fmt.Sprintf(
        "You have extended your reservation on \"*%v*\". It now expires" +
        " in *%v*",
        resource,
        reservation.RemainingTimeToString())

    return response, true

}

/*
Run this locally with:

curl -XPOST \
     -H "Content-Type: application/json" \
     -d @example/cancel \
     http://localhost:8080/slack/commands/reservations

*/
func handleCommandCancel(slack_request SlackRequest) (SlackResponse, bool) {

    command  := slack_request.FormattedSubcommand()
    response := SlackResponse{}

    // Extract data from command
    matches     := subcmd_cancel_regex.FindStringSubmatch(command)
    resource    := matches[1]

    // Check that resource is valid
    // Delete() below checks for this, but we want to do it sooner so we
    // can return the appropriate message to the user
    if !IsValidResource(resource) {
        response.Text = unknownResourceText(resource)
        return response, true
    }

    // Find all reservations
    reservations, err := NewReservations()
    if err != nil {
        log.Error(err)
        return response, false
    }

    // Ensure an active reservation exists for this reousrce and user.
    reservation := reservations.FindByResource(resource)
    if !reservation.IsPresent() ||
            !reservation.IsActive() ||
            slack_request.UserName != reservation.User {
        response.Text = fmt.Sprintf(
                "You don't have any reservation on \"*%v*\" to cancel\n\n" +
                    "Type `/reservations list` to list current reservations",
                resource)

        return response, true
    }

    // Delete
    // No need to check explicitly for `isInvalidResourceError()` since
    // that's already done manually above
    err = reservations.Delete(resource)
    if err != nil {
        log.Error(err)
        return response, false
    }

    // Save to file
    err = reservations.WriteToFile()
    if err != nil { log.Error(err); return response, false }

    // Construct a response for the user
    response.Text = fmt.Sprintf(
        "Your reservation on \"*%v*\" has been cancelled",
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

func unknownResourceText(resource string) string {

    return fmt.Sprintf(
        "I don't what \"*%v*\" is. Did you misspell it?\nValid resources: %v",
        resource,
        ListOfResources(),
    )

}

func isInvalidResourceError(err error) bool {

    return regexp.MustCompile("Invalid Resource").MatchString(err.Error())

}

func buildInvalidResponse(w http.ResponseWriter) {

    response := SlackResponse{
        Text: "Yikes, that looks like an invalid request!",
    }
    buildResponse(response, w)

}

func buildErrorResponse(w http.ResponseWriter) {

    response := SlackResponse{
        Text: "Sorry, I couldn't understand your request.\nType " +
            "/reservations help for more info",
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
