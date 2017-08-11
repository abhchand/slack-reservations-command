package main

import(
    "fmt"
    "math"
    "time"
)

const(
    SECS_PER_DAY    = 60 * 60 * 24
    SECS_PER_HOUR   = 60 * 60
    SECS_PER_MINUTE = 60
)

type Reservation struct {
    User      string    `json:"user"`
    EndAt     time.Time `json:"end_at"`
}

func (r Reservation) IsPresent() bool {

    return r != Reservation{}

}

func (r Reservation) IsActive() bool {
    return !r.EndAt.IsZero() && r.EndAt.After(time.Now())
}

func (r Reservation) RemainingTimeToString() string {

    if !r.IsActive() { return r.formatDuration(0.0, "minute") }

    // When someone reserves a resource for "3 hours" they get a response
    // with "2 hours, 59 minutes".
    // This is because making the web request takes some non-zero amount of
    // time and the logic rounds *down*
    // Just add 3.0 seconds to work around this problem. It's small enough
    // to not matter, but will keep the output correctly formatted
    // Slack requests also time out after 3 seconds (3000ms) so this should
    // be the maximum time needed
    s := r.EndAt.Sub(time.Now()).Seconds() + 3.0

    switch {

    case s >= SECS_PER_DAY:
        days := r.formatDuration(s / SECS_PER_DAY, "day")
        hours := r.formatDuration(
            (s - (math.Floor(s / SECS_PER_DAY) * SECS_PER_DAY)) / SECS_PER_HOUR,
            "hour")

        return fmt.Sprintf("%v, %v", days, hours)

    case s >= SECS_PER_HOUR:
        hours := r.formatDuration(s / SECS_PER_HOUR, "hour")
        mins := r.formatDuration(
            (s - (math.Floor(s / SECS_PER_HOUR) * SECS_PER_HOUR)) / SECS_PER_MINUTE,
            "minute")

        return fmt.Sprintf("%v, %v", hours, mins)

    default:
        return r.formatDuration(s / SECS_PER_MINUTE, "minute")

    }

}

func (r Reservation) formatDuration(duration float64, unit string) string {

    value := int(math.Floor(duration))

    // This is not proper I81n, but luckily pluralization of
    // these words in english is just adding an 's'. Refactor later as needed
    if value != 1 { unit = unit + "s" }

    return fmt.Sprintf("%v %v", value, unit)

}
