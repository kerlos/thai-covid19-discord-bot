package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
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
	retryCount := 0
	var err error
	for {
		if retryCount > 3 {
			return nil, err
		} else if retryCount > 0 {
			time.Sleep(10 * time.Second)
		}
		cl := http.Client{}

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			retryCount++
			continue
		}

		res, err := cl.Do(req)
		if err != nil {
			retryCount++
			continue
		}

		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			retryCount++
			continue
		}

		data := covidData{}
		err = json.Unmarshal(body, &data)
		//sometime api return empty content
		if err != nil {
			retryCount++
			continue
		}
		return &data, nil
	}
}
