package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestNewReservations(t *testing.T) {

	//
	// Setup
	//

	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "production, staging")

	reservations_file = reservations_file + ".test"

	t.Run("Success", func(t *testing.T) {

		expected := Reservations{
			"production": Reservation{
				User: "foo", EndAt: time.Now().AddDate(0, 0, 1)},
			"staging": Reservation{
				User: "foo", EndAt: time.Now().AddDate(0, 0, 2)},
		}

		body, err := json.Marshal(expected)
		if err != nil {
			panic(err)
		}

		err = writeToReservationsFile(string(body))
		if err != nil {
			t.Error("Expected no error writing to file. Got", err)
		}

		actual, err := NewReservations()
		if err != nil {
			t.Error("Error while calling NewReservations():", err)
		}

		if len(actual) != len(expected) {
			t.Error(
				"expected length", len(expected),
				"got length", len(actual),
			)
		}

		for key, value := range expected {
			e := value
			a := actual[key]

			if a != e {
				t.Error(
					"expected", e,
					"got", a,
				)
			}
		}
	})

	t.Run("FailToReadFromFile", func(t *testing.T) {

		err := writeToReservationsFile("{}")
		if err != nil {
			t.Error("Expected no error writing to file. Got", err)
		}

		// Delete file
		err = os.Remove(reservations_file)
		if err != nil {
			panic(err)
		}

		_, err = NewReservations()

		actual := err.Error()
		expected := fmt.Sprintf(
			"open %v: no such file or directory", reservations_file)

		if actual != expected {
			t.Error(
				"expected", expected,
				"got", actual,
			)
		}
	})

	t.Run("FailToParseJsonData", func(t *testing.T) {

		err := writeToReservationsFile("something that's not json")
		if err != nil {
			t.Error("Expected no error writing to file. Got", err)
		}

		_, err = NewReservations()

		actual := err.Error()
		expected := "invalid character 's' looking for beginning of value"

		if actual != expected {
			t.Error(
				"expected", expected,
				"got", actual,
			)
		}
	})

}

func TestWriteToFile(t *testing.T) {

	//
	// Setup
	//

	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "production, staging")

	reservations_file = reservations_file + ".test"

	dateFormat := "2006-01-02T15:04:05.000000000-07:00"
	timestamp := "2017-08-11T17:48:37.556835687-04:00"

	t.Run("Success", func(t *testing.T) {

		endAt, _ := time.Parse(dateFormat, timestamp)

		reservations := Reservations{
			"staging": Reservation{User: "foo", EndAt: endAt},
		}

		err := reservations.WriteToFile()
		if err != nil {
			t.Error("Error while calling WriteFile():", err)
		}

		expected :=
			fmt.Sprintf(
				"{\"staging\":{\"user\":\"%v\",\"end_at\":\"%v\"}}",
				reservations["staging"].User,
				timestamp,
			)

		body, err := ioutil.ReadFile(reservations_file)
		if err != nil {
			t.Error("Error reading from file", err)
		}
		actual := string(body)

		if actual != expected {
			t.Error(
				"expected", expected,
				"got", actual,
			)
		}

	})

	t.Run("FailToParseJsonData", func(t *testing.T) {
		// Placeholder - how to test an error while marshalling out data
		// from a struct?
	})

	t.Run("FailToWriteToFile", func(t *testing.T) {

		// Delete file
		// _, err := os.Stat(reservations_file)
		// if !os.IsNotExist(err) {
		//     err := os.Remove(reservations_file)
		//     if err != nil { panic(err) }
		// }

		// reservations := Reservations{
		//     "production": Reservation{User:"foo", EndAt:time.Now()},
		// }
		// err = reservations.WriteToFile()

		// actual := err.Error()
		// expected := fmt.Sprintf(
		//     "open %v: no such file or directory", reservations_file)

		// if actual != expected {
		//     t.Error(
		//         "expected", expected,
		//         "got", actual,
		//     )
		// }

	})

}

func TestFindByResource(t *testing.T) {

	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "production, staging")

	reservations_file = reservations_file + ".test"

	r1 := Reservation{User: "abc", EndAt: time.Now().AddDate(0, 0, 1)}
	r2 := Reservation{User: "def", EndAt: time.Now().AddDate(0, 0, 1)}
	reservations := Reservations{"production": r1, "staging": r2}

	test_cases := map[string]Reservation{
		"production": r1,
		"foo":        Reservation{},
	}

	for resource, expected := range test_cases {
		actual := reservations.FindByResource(resource)

		if actual != expected {
			t.Error(
				"expected", expected,
				"got", actual,
			)
		}
	}
}

func TestUpsert(t *testing.T) {

	// Setup

	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "production, staging")

	reservations_file = reservations_file + ".test"

	t.Run("Success", func(t *testing.T) {

		r1 := Reservation{User: "abc", EndAt: time.Now().AddDate(0, 0, 1)}
		r2 := Reservation{User: "def", EndAt: time.Now().AddDate(0, 0, 1)}
		r3 := Reservation{User: "ghi", EndAt: time.Now().AddDate(0, 0, 1)}

		reservations := Reservations{"production": r1, "staging": r2}

		err := reservations.Upsert("staging", r3)

		if err != nil {
			t.Error("Expected no error, got", err)
		}

		if actual := reservations.FindByResource("production"); actual != r1 {
			t.Error("expected", r1, "got", actual)
		}

		if actual := reservations.FindByResource("staging"); actual != r3 {
			t.Error("expected", r3, "got", actual)
		}

	})

	t.Run("InvalidResource", func(t *testing.T) {

		r1 := Reservation{User: "abc", EndAt: time.Now().AddDate(0, 0, 1)}
		r2 := Reservation{User: "def", EndAt: time.Now().AddDate(0, 0, 1)}
		r3 := Reservation{User: "ghi", EndAt: time.Now().AddDate(0, 0, 1)}

		reservations := Reservations{"production": r1, "staging": r2}

		err := reservations.Upsert("foo", r3)

		if err == nil ||
			!regexp.MustCompile("Invalid Resource").MatchString(err.Error()) {
			t.Error("expect error message /Invalid Resource/, got", err)
		}

		// Check that existing values remain unchanged

		if actual := reservations.FindByResource("production"); actual != r1 {
			t.Error("expected", r1, "got", actual)
		}

		if actual := reservations.FindByResource("staging"); actual != r2 {
			t.Error("expected", r2, "got", actual)
		}

	})
}

func TestDelete(t *testing.T) {

	// Setup

	old_env := os.Getenv("RESOURCES")
	defer os.Setenv("RESOURCES", old_env)
	os.Setenv("RESOURCES", "production, staging")

	reservations_file = reservations_file + ".test"

	t.Run("Success", func(t *testing.T) {

		r1 := Reservation{User: "abc", EndAt: time.Now().AddDate(0, 0, 1)}
		r2 := Reservation{User: "def", EndAt: time.Now().AddDate(0, 0, 1)}
		zero_value := Reservation{}

		reservations := Reservations{"production": r1, "staging": r2}

		err := reservations.Delete("staging")

		if err != nil {
			t.Error("Expected no error, got", err)
		}

		if actual := reservations.FindByResource("production"); actual != r1 {
			t.Error("expected", r1, "got", actual)
		}

		if actual := reservations.FindByResource("staging"); actual != zero_value {
			t.Error("expected", zero_value, "got", actual)
		}

	})

	t.Run("InvalidResource", func(t *testing.T) {

		r1 := Reservation{User: "abc", EndAt: time.Now().AddDate(0, 0, 1)}
		r2 := Reservation{User: "def", EndAt: time.Now().AddDate(0, 0, 1)}

		reservations := Reservations{"production": r1, "staging": r2}

		err := reservations.Delete("foo")

		if err == nil ||
			!regexp.MustCompile("Invalid Resource").MatchString(err.Error()) {
			t.Error("expect error message /Invalid Resource/, got", err)
		}

		// Check that existing values remain unchanged

		if actual := reservations.FindByResource("production"); actual != r1 {
			t.Error("expected", r1, "got", actual)
		}

		if actual := reservations.FindByResource("staging"); actual != r2 {
			t.Error("expected", r2, "got", actual)
		}

	})

	t.Run("NoValueForResource", func(t *testing.T) {

		r1 := Reservation{User: "abc", EndAt: time.Now().AddDate(0, 0, 1)}
		zero_value := Reservation{}

		reservations := Reservations{"production": r1}

		err := reservations.Delete("staging")

		if err != nil {
			t.Error("Expected no error, got", err)
		}

		// Check that existing values remain unchanged

		if actual := reservations.FindByResource("production"); actual != r1 {
			t.Error("expected", r1, "got", actual)
		}

		if actual := reservations.FindByResource("staging"); actual != zero_value {
			t.Error("expected", zero_value, "got", actual)
		}

	})
}

func writeToReservationsFile(body string) error {

	var err error

	// Create directory if it does not exist
	_, err = os.Stat(reservations_dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(reservations_dir, 0775)

		if err != nil {
			log.Debugf("Error creating directory %v", reservations_dir)
			return err
		}
	}

	// Overwrite existing test file
	err = ioutil.WriteFile(reservations_file, []byte(body), 0755)
	if err != nil {
		return err
	}

	return nil

}
