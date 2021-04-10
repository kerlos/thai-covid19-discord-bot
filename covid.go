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

	province struct {
		Slug  string `json:"slug"`
		Title string `json:"title"`
	}

	provinceDataLatest struct {
		LastUpdated string `json:"lastUpdated"`
		URL         string `json:"url"`
	}

	provinceDataResponse struct {
		LastUpdated string         `json:"lastUpdated"`
		Data        []provinceData `json:"data"`
	}

	provinceData struct {
		Slug          string `json:"slug"`
		Title         string `json:"title"`
		CurrentStatus struct {
			Accumulate                 int `json:"accumulate"`
			New                        int `json:"new"`
			InfectionLevelByRule       int `json:"infectionLevelByRule"`
			InfectionLevelByPercentile int `json:"infectionLevelByPercentile"`
		} `json:"currentStatus"`
		Rank int
	}
)

const apiURL = "https://disease.sh/v3/covid-19"
const provinceDataURL = "https://s.isanook.com/an/0/covid-19/static/data/thailand/accumulate"

var checkResults []checkResult
var provinces map[string]string

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
func getProvinceDataLatest() (*provinceDataLatest, error) {
	retryCount := 0
	var err error
	for {
		if retryCount > 3 {
			return nil, err
		} else if retryCount > 0 {
			time.Sleep(10 * time.Second)
		}
		cl := http.Client{}
		dt := time.Now().Unix()

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/latest.json?%v", provinceDataURL, dt), nil)
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

		data := provinceDataLatest{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			retryCount++
			continue
		}

		return &data, nil
	}
}

func getProvinceData(date string) (data *provinceDataResponse, err error) {
	ck := "p_" + date
	cv, ok := ca.Get(ck)
	if ok {
		err := json.Unmarshal(cv.([]byte), &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	retryCount := 0
	for {
		if retryCount > 3 {
			return nil, err
		} else if retryCount > 0 {
			time.Sleep(10 * time.Second)
		}
		cl := http.Client{}
		dt := time.Now().Unix()

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s.json?%v", provinceDataURL, date, dt), nil)
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

		err = json.Unmarshal(body, &data)
		if err != nil {
			retryCount++
			continue
		}
		if len(data.Data) == 0 {
			retryCount++
			continue
		}

		storeData, err := json.Marshal(&data)
		if err != nil {
			return nil, err
		}
		ca.Set(ck, storeData, time.Hour*36)
		return data, nil
	}
}
