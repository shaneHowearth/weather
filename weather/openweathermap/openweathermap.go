package openweathermap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// OpenWeather -
type OpenWeather struct {
	url string
}

// NewOpenWeather -
func NewOpenWeather() (*OpenWeather, error) {
	return &OpenWeather{
		url: "http://api.openweathermap.org/data/2.5/weather",
	}, nil
}

// Data -
// DAO to receive data from upstream service
// Note - this endpoint defaults to Metric
// TODO - Not all of this will be used, so there's no need to allocate momory.
// (so trim it down!)
type Data struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

// DTO
type response struct {
	Temperature   float64
	Feelslike     float64
	Pressure      int
	Humidity      int
	WindSpeed     float64
	WindDirection int
	Visibility    int
}

// Allow http.Get to be faked in unit tests
var httpGet = http.Get

// allow ioutil.ReadAll to be faked for tests
var ioutilReadAll = ioutil.ReadAll

// allow json.Unmarshal to be faked for tests
var jsonUnmarshal = json.Unmarshal

// GetWeather -
// ignore the linter warning about returning an unexported type
// nolint:revive
func (ow *OpenWeather) GetWeather(city, appID string) (response, error) {
	if city == "" {
		return response{}, fmt.Errorf("city is required")
	}
	if appID == "" {
		return response{}, fmt.Errorf("appID is required")
	}
	// build query string
	query := ow.url + "?q=" + city + "&appid=" + appID + "&units=metric"

	// Make call to server
	resp, err := httpGet(query)
	if err != nil {
		return response{}, fmt.Errorf("getWeather: http.Get error %w", err)
	}
	defer resp.Body.Close()

	// Check that the server is happy with out request
	if resp.StatusCode != http.StatusOK {
		return response{}, fmt.Errorf("getWeather: got bad status %d", resp.StatusCode)
	}
	body, err := ioutilReadAll(resp.Body)
	if err != nil {
		return response{}, fmt.Errorf("getWeather: reading response error %w", err)
	}

	a := Data{}
	if err := jsonUnmarshal(body, &a); err != nil {
		return response{}, fmt.Errorf("getWeather: unmarshalling response error %w", err)
	}

	return response{
		Temperature:   a.Main.Temp,
		Feelslike:     a.Main.FeelsLike,
		Pressure:      a.Main.Pressure,
		Humidity:      a.Main.Humidity,
		WindSpeed:     a.Wind.Speed,
		WindDirection: a.Wind.Deg,
		Visibility:    a.Visibility,
	}, nil
}
