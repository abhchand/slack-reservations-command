package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

	return reservations, nil

}

func (r Reservations) WriteToFile() error {

	log.Debugf("Writing to reservations file %v", reservations_file)

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

func (r Reservations) Upsert(resource string, reservation Reservation) error {

	if !IsValidResource(resource) {
		return errors.New(fmt.Sprintf("Invalid Resource: %v", resource))
	}

	r[resource] = reservation
	return nil

}

func (r Reservations) Delete(resource string) error {

	if !IsValidResource(resource) {
		return errors.New(fmt.Sprintf("Invalid Resource: %v", resource))
	}

	if (r[resource] != Reservation{}) {
		delete(r, resource)
	}
	return nil

}
