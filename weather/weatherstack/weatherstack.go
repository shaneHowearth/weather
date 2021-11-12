package weatherstack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// WeatherStack -
type WeatherStack struct {
	url       string
	accessKey string
}

// NewWeatherStack -
func NewWeatherStack(accessKey string) (*WeatherStack, error) {
	if accessKey == "" {
		return nil, fmt.Errorf("accessKey is required")
	}
	return &WeatherStack{
		url:       "http://api.weatherstack.com/current",
		accessKey: accessKey,
	}, nil
}

// Data -
// Data Access Object
type Data struct {
	Current struct {
		Temperature int `json:"temperature"`
		WindSpeed   int `json:"wind_speed"`
	} `json:"current"`
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
func (ws *WeatherStack) GetWeather(city string) (response, error) {
	if city == "" {
		return response{}, fmt.Errorf("city is required")
	}

	// build query string - note units are hardcoded to metric
	query := ws.url + "?query=" + city + "&access_key=" + ws.accessKey + "&units=metric"

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
		Temperature: float64(a.Current.Temperature),
		WindSpeed:   float64(a.Current.WindSpeed),
	}, nil
}
