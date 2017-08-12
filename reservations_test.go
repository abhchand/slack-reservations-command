package main

import(
    "encoding/json"
    "fmt"
    "path/filepath"
    "io/ioutil"
    "os"
    "testing"
    "time"
)


func TestNewReservations(t *testing.T) {

    //
    // Setup
    //

    old_env := os.Getenv("RESOURCES")
    defer os.Setenv("RESOURCES", old_env)
    os.Setenv("RESOURCES", "development, staging")

    reservations_file = filepath.Join(
        reservations_dir, "reservations-test.json")


    t.Run("Success", func(t *testing.T) {

        expected := Reservations{
            "development": Reservation{
                User:"foo", EndAt:time.Now().AddDate(0, 0, 1)},
            "staging": Reservation{
                User:"foo", EndAt:time.Now().AddDate(0, 0, 2)},
        }

        body, err := json.Marshal(expected)
        if err != nil { panic(err) }
        writeToReservationsFile(string(body))

        actual, err := NewReservations()
        if err != nil { t.Error("Error while calling NewReservations():", err) }

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

        writeToReservationsFile("{}")

        // Delete file
        err := os.Remove(reservations_file)
        if err != nil { panic(err) }

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

        writeToReservationsFile("something that's not json")

        _, err := NewReservations()

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
    os.Setenv("RESOURCES", "development, staging")

    reservations_file = filepath.Join(
        reservations_dir, "reservations-test.json")

    dateFormat := "2006-01-02T15:04:05.000000000-07:00"
    timestamp  := "2017-08-11T17:48:37.556835687-04:00"

    t.Run("Success", func(t *testing.T) {

        endAt, _ := time.Parse(dateFormat, timestamp)

        reservations := Reservations{
            "staging": Reservation{User:"foo", EndAt:endAt},
        }

        err := reservations.WriteToFile()
        if err != nil { t.Error("Error while calling WriteFile():", err) }

        expected :=
            fmt.Sprintf(
                "{\"staging\":{\"user\":\"%v\",\"end_at\":\"%v\"}}",
                reservations["staging"].User,
                timestamp,
            )

        body, err := ioutil.ReadFile(reservations_file)
        if err != nil { t.Error("Error reading from file", err) }
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
        //     "development": Reservation{User:"foo", EndAt:time.Now()},
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

func writeToReservationsFile(body string) error {

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

    // Overwrite existing test file
    err = ioutil.WriteFile(reservations_file, []byte(body), 0755)
    if err != nil { return err }

    return nil

}
