package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type covidData struct {
	Confirmed       int    `json:"Confirmed"`
	Recovered       int    `json:"Recovered"`
	Hospitalized    int    `json:"Hospitalized"`
	Deaths          int    `json:"Deaths"`
	NewConfirmed    int    `json:"NewConfirmed"`
	NewRecovered    int    `json:"NewRecovered"`
	NewHospitalized int    `json:"NewHospitalized"`
	NewDeaths       int    `json:"NewDeaths"`
	UpdateDate      string `json:"UpdateDate"`
	Source          string `json:"Source"`
	DevBy           string `json:"DevBy"`
	SeverBy         string `json:"SeverBy"`
}

const apiURL = "https://covid19.th-stat.com/api/open/today"

func getData() (*covidData, error) {

	cl := http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	data := covidData{}
	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		return nil, err
	}

	return &data, nil
}
