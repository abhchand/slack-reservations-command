package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "strings"
)

type Reservations map[string]Reservation

func NewReservations() (Reservations, error) {

    log.Debugf("Reading reservations file %v", reservations_file)

    var reservations Reservations

    // Read from file
    body, err := ioutil.ReadFile(reservations_file)
    if err != nil {
        log.Error("Could not read from file")
        return reservations, err
    }

    // Parse JSON data
    err = json.Unmarshal(body, &reservations)
    if err != nil {
        log.Error("Could not unmarshal JSON data")
        return reservations, err
    }

    // Check validity of data
    err = reservations.validate()
    if err != nil {
        log.Errorf(err.Error())
        return reservations, err
    }

    return reservations, nil

}

func (r Reservations) WriteToFile() error {

    log.Debugf("Writing to reservations file %v", reservations_file)

    // Check validity of data
    err := r.validate()
    if err != nil {
        log.Errorf(err.Error())
        return err
    }

    // Create JSON data
    body, err := json.Marshal(r)
    if err != nil {
        log.Error("Could not marshal JSON data")
        return err
    }

    // Write to file
    err = ioutil.WriteFile(reservations_file, body, 0755)
    if err != nil {
        log.Error("Could not write to file")
        return err
    }

    return nil

}

func (r Reservations) FindByResource(resource string) Reservation {

    return r[resource]

}

func (r Reservations) Upsert(resource string, reservation Reservation) {

    r[resource] = reservation

}

func (r Reservations) Delete(resource string) {

    if (r[resource] != Reservation{}) { delete(r, resource) }

}

func (r Reservations) validate() error {

    resources := ListOfResources()

    // Initialize count of resources
    freq_counts := make(map[string]int)
    for _, i := range resources { freq_counts[strings.ToLower(i)] = 0 }

    // Check that there are no duplicate resources
    for resource, _ := range r { freq_counts[resource] += 1 }
    for resource, count := range freq_counts {
        if count > 1 {
            err_txt := fmt.Sprintf(
                "Duplicate value %v in reservations file", resource)
            log.Debugf(err_txt)
            return errors.New(err_txt)
        }
    }

    // Check that all resources are in the valid list of allowable resources
    for listed_resource, _ := range r {
        present := false
        for _, allowable_resource := range resources {
            if listed_resource == allowable_resource { present = true; break }
        }

        if !present {
            err_txt := fmt.Sprintf(
                "Resource %v is not a valid resource", listed_resource)
            log.Debugf(err_txt)
            return errors.New(err_txt)
        }
    }

    return nil

}
