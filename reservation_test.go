package main

import (
	"testing"
	"time"
)

func TestIsPresent(t *testing.T) {

	present := Reservation{User: "foo", EndAt: time.Now()}
	blank := Reservation{}

	data := map[Reservation]bool{
		present: true,
		blank:   false,
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

	active := Reservation{User: "foo", EndAt: time.Now().AddDate(0, 0, 1)}
	inactive := Reservation{User: "foo", EndAt: time.Now().AddDate(0, 0, -1)}

	data := map[Reservation]bool{
		active:   true,
		inactive: false,
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

	// Re-define for readability below
	secs := 1
	mins := SECS_PER_MINUTE
	hrs := SECS_PER_HOUR
	days := SECS_PER_DAY

	test_cases := map[int]string{

		// Days remaining
		(2*days + 2*hrs + 0*mins + 0*secs): "2 days, 2 hours",
		(2*days + 1*hrs + 0*mins + 0*secs): "2 days, 1 hour",
		(2*days + 0*hrs + 0*mins + 0*secs): "2 days, 0 hours",
		(1*days + 2*hrs + 0*mins + 0*secs): "1 day, 2 hours",
		(1*days + 1*hrs + 0*mins + 0*secs): "1 day, 1 hour",
		(1*days + 0*hrs + 0*mins + 0*secs): "1 day, 0 hours",

		// Hours remaining
		(0*days + 2*hrs + 30*mins + 0*secs): "2 hours, 30 minutes",
		(0*days + 2*hrs + 1*mins + 0*secs):  "2 hours, 1 minute",
		(0*days + 2*hrs + 0*mins + 0*secs):  "2 hours, 0 minutes",
		(0*days + 1*hrs + 30*mins + 0*secs): "1 hour, 30 minutes",
		(0*days + 1*hrs + 1*mins + 0*secs):  "1 hour, 1 minute",
		(0*days + 1*hrs + 0*mins + 0*secs):  "1 hour, 0 minutes",

		// Minutes remaining
		(0*days + 0*hrs + 30*mins + 0*secs): "30 minutes",
		(0*days + 0*hrs + 1*mins + 0*secs):  "1 minute",
		(0*days + 0*hrs + 0*mins + 0*secs):  "0 minutes",

		// Rounding up
		(0*days + 0*hrs + 0*mins + 1*secs):   "1 minute",
		(0*days + 0*hrs + 59*mins + 1*secs):  "1 hour, 0 minutes",
		(0*days + 23*hrs + 59*mins + 1*secs): "1 day, 0 hours",
	}

	for seconds, expected := range test_cases {

		endAt := time.Now().Local().Add(time.Second * time.Duration(seconds))

		actual := Reservation{User: "foo", EndAt: endAt}.RemainingTimeToString()

		if actual != expected {
			t.Error(
				"expected", expected,
				"got", actual,
			)
		}
	}

}
