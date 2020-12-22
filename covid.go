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
		Cases          int   `json:"cases"`
		Recovered      int   `json:"recovered"`
		Active         int   `json:"active"`
		Deaths         int   `json:"deaths"`
		TodayCases     int   `json:"todayCases"`
		TodayRecovered int   `json:"todayRecovered"`
		TodayDeaths    int   `json:"todayDeaths"`
		Updated        int64 `json:"updated"`
	}
	chartData struct {
		Timeline struct {
			Cases     map[string]int `json:"cases"`
			Deaths    map[string]int `json:"deaths"`
			Recovered map[string]int `json:"recovered"`
		} `json:"timeline"`
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

const apiURL = "https://disease.sh/v3/covid-19"

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

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/countries/thailand", apiURL), nil)
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
		if data.Cases == 0 {
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

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/historical/th?lastdays=30", apiURL), nil)
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
