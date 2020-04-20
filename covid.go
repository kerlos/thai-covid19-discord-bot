package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type (
	covidData struct {
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
	chartData struct {
		UpdateDate string `json:"UpdateDate"`
		Source     string `json:"Source"`
		DevBy      string `json:"DevBy"`
		SeverBy    string `json:"SeverBy"`
		Data       []struct {
			Date            string `json:"Date"`
			NewConfirmed    int    `json:"NewConfirmed"`
			NewRecovered    int    `json:"NewRecovered"`
			NewHospitalized int    `json:"NewHospitalized"`
			NewDeaths       int    `json:"NewDeaths"`
			Confirmed       int    `json:"Confirmed"`
			Recovered       int    `json:"Recovered"`
			Hospitalized    int    `json:"Hospitalized"`
			Deaths          int    `json:"Deaths"`
		} `json:"Data"`
	}

	checkResult struct {
		Index             int    `json:"index"`
		Fever             int    `json:"fever"`
		OneURISymp        int    `json:"one_uri_symp"`
		TravelRiskCountry int    `json:"travel_risk_country"`
		Covid19Contact    int    `json:"covid19_contact"`
		CloseRiskCountry  int    `json:"close_risk_country"`
		CloseRiskLocation int    `json:"close_risk_location"`
		IntContact        int    `json:"int_contact"`
		MedProf           int    `json:"med_prof"`
		CloseCon          int    `json:"close_con"`
		RiskLevel         int    `json:"risk_level"`
		GenAction         string `json:"gen_action"`
		SpecAction        string `json:"spec_action"`
	}
)

const apiURL = "https://covid19.th-stat.com/api/open"

var checkResults []checkResult

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

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/today", apiURL), nil)
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

func getChartData() (*chartData, error) {
	retryCount := 0
	var err error
	for {
		if retryCount > 3 {
			return nil, err
		} else if retryCount > 0 {
			time.Sleep(10 * time.Second)
		}
		cl := http.Client{}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/timeline", apiURL), nil)
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

		data := chartData{}
		err = json.Unmarshal(body, &data)
		//sometime api return empty content
		if err != nil {
			retryCount++
			continue
		}
		return &data, nil
	}
}
