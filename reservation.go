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

    // Get the number of seconds elapse, but round up to the nearest minute
    // first
    s := math.Ceil(r.EndAt.Sub(time.Now()).Minutes()) * SECS_PER_MINUTE

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
