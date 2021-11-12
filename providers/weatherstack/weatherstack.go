package weatherstack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

// Allow http.Get to be faked in unit tests
var httpGet = http.Get

// allow ioutil.ReadAll to be faked for tests
var ioutilReadAll = ioutil.ReadAll

// allow json.Unmarshal to be faked for tests
var jsonUnmarshal = json.Unmarshal

// GetWeather -
// ignore the linter warning about returning an unexported type
// nolint:revive
func (ws *WeatherStack) GetWeather(city string) (struct {
	Temperature, WindSpeed float64
}, error) {
	if city == "" {
		return struct{ Temperature, WindSpeed float64 }{}, fmt.Errorf("city is required")
	}
	wsCity, ok := ws.getCity(city)
	if !ok {
		return struct{ Temperature, WindSpeed float64 }{}, fmt.Errorf("%q is an unknown city for this provider", city)
	}

	// build query string - note units are hardcoded to metric
	query := ws.url + "?query=" + wsCity + "&access_key=" + ws.accessKey + "&units=metric"

	// Make call to server
	resp, err := httpGet(query)
	if err != nil {
		return struct{ Temperature, WindSpeed float64 }{}, fmt.Errorf("getWeather: http.Get error %w", err)
	}
	defer resp.Body.Close()

	// Check that the server is happy with out request
	if resp.StatusCode != http.StatusOK {
		return struct{ Temperature, WindSpeed float64 }{}, fmt.Errorf("getWeather: got bad status %d", resp.StatusCode)
	}
	body, err := ioutilReadAll(resp.Body)
	if err != nil {
		return struct{ Temperature, WindSpeed float64 }{}, fmt.Errorf("getWeather: reading response error %w", err)
	}

	a := Data{}
	if err := jsonUnmarshal(body, &a); err != nil {
		return struct{ Temperature, WindSpeed float64 }{}, fmt.Errorf("getWeather: unmarshalling response error %w", err)
	}

	return struct {
		Temperature, WindSpeed float64
	}{
		Temperature: float64(a.Current.Temperature),
		WindSpeed:   float64(a.Current.WindSpeed),
	}, nil
}

// Simple datastore for city name conversion
// this could easily be an external cache with the orchestrators knowledge of
// countries/provinces or states/cities as keys and this providers city list as
// values
var cities = map[string]string{"melbourne": "Melbourne"}

func (ws *WeatherStack) getCity(city string) (string, bool) {
	c, ok := cities[strings.ToLower(strings.TrimSpace(city))]
	return c, ok
}
