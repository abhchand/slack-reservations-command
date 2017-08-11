package main

import(
  "testing"
  "time"
)

func TestIsPresent(t *testing.T) {

    present := Reservation{User:"foo", EndAt:time.Now()}
    blank   := Reservation{}

    data := map[Reservation]bool{
        present:    true,
        blank:      false,
    }

    for r, expected := range data {
        actual := r.IsPresent()

        if actual != expected {
            t.Error(
                "expected", expected,
                "got", actual,
            )
        }
    }

}

func TestIsActive(t *testing.T) {

    active      := Reservation{User:"foo", EndAt:time.Now().AddDate(0, 0, 1)}
    inactive    := Reservation{User:"foo", EndAt:time.Now().AddDate(0, 0, -1)}

    data := map[Reservation]bool{
        active:     true,
        inactive:   false,
    }

    for r, expected := range data {
        actual := r.IsActive()

        if actual != expected {
            t.Error(
                "expected", expected,
                "got", actual,
            )
        }
    }

}

func TestRemainingTimeToString(t *testing.T) {

    data := map[int]string{
        // Minutes: Expected result
        (2*1440 + 120): "2 days, 2 hours",
        (2*1440 + 60):  "2 days, 1 hour",
        (2*1440 + 0):   "2 days, 0 hours",
        (1*1440 + 120): "1 day, 2 hours",
        (1*1440 + 60):  "1 day, 1 hour",
        (1*1440 + 0):   "1 day, 0 hours",
        (2*60 + 30):    "2 hours, 30 minutes",
        (2*60 + 1):     "2 hours, 1 minute",
        (2*60 + 0):     "2 hours, 0 minutes",
        (1*60 + 30):    "1 hour, 30 minutes",
        (1*60 + 1):     "1 hour, 1 minute",
        (1*60 + 0):     "1 hour, 0 minutes",
        (30):           "30 minutes",
        (1):            "1 minute",
        (0):            "0 minutes",
    }

    for minutes, expected := range data {

        endAt := time.Now().Local().Add(time.Minute * time.Duration(minutes))

        actual := Reservation{User:"foo", EndAt:endAt}.RemainingTimeToString()

        if actual != expected {
            t.Error(
                "expected", expected,
                "got", actual,
            )
        }
    }

}
