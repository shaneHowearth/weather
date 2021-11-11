package openweathermap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// OpenWeather -
type OpenWeather struct {
	url   string
	appID string
}

// NewOpenWeather -
func NewOpenWeather(appID string) (*OpenWeather, error) {
	if appID == "" {
		return nil, fmt.Errorf("appID is required")
	}
	return &OpenWeather{
		url:   "http://api.openweathermap.org/data/2.5/weather",
		appID: appID,
	}, nil
}

// Data -
// DAO to receive data from upstream service
type Data struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

// DTO
type response struct {
	Temperature float64
	WindSpeed   float64
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
func (ow *OpenWeather) GetWeather(city string) (response, error) {
	if city == "" {
		return response{}, fmt.Errorf("city is required")
	}
	// build query string
	query := ow.url + "?q=" + city + "&appid=" + ow.appID + "&units=metric"

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
		Temperature: a.Main.Temp,
		WindSpeed:   a.Wind.Speed,
	}, nil
}
