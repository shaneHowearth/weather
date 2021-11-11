package weatherstack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// WeatherStack -
type WeatherStack struct {
	url string
}

// NewWeatherStack -
func NewWeatherStack() (*WeatherStack, error) {
	return &WeatherStack{
		url: "http://api.weatherstack.com/current",
	}, nil
}

// Data -
type Data struct {
	Request struct {
		Type     string `json:"type"`
		Query    string `json:"query"`
		Language string `json:"language"`
		Unit     string `json:"unit"`
	} `json:"request"`
	Location struct {
		Name           string `json:"name"`
		Country        string `json:"country"`
		Region         string `json:"region"`
		Lat            string `json:"lat"`
		Lon            string `json:"lon"`
		TimezoneID     string `json:"timezone_id"`
		Localtime      string `json:"localtime"`
		LocaltimeEpoch int    `json:"localtime_epoch"`
		UtcOffset      string `json:"utc_offset"`
	} `json:"location"`
	Current struct {
		ObservationTime     string   `json:"observation_time"`
		Temperature         int      `json:"temperature"`
		WeatherCode         int      `json:"weather_code"`
		WeatherIcons        []string `json:"weather_icons"`
		WeatherDescriptions []string `json:"weather_descriptions"`
		WindSpeed           int      `json:"wind_speed"`
		WindDegree          int      `json:"wind_degree"`
		WindDir             string   `json:"wind_dir"`
		Pressure            int      `json:"pressure"`
		Precip              int      `json:"precip"`
		Humidity            int      `json:"humidity"`
		Cloudcover          int      `json:"cloudcover"`
		FeelsLike           int      `json:"feelslike"`
		UvIndex             int      `json:"uv_index"`
		Visibility          int      `json:"visibility"`
		IsDay               string   `json:"is_day"`
	} `json:"current"`
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
func (ow *WeatherStack) GetWeather(city, access_key string) (response, error) {
	if city == "" {
		return response{}, fmt.Errorf("city is required")
	}
	if access_key == "" {
		return response{}, fmt.Errorf("access_key is required")
	}

	// build query string - note units are hardcoded to metric
	query := ow.url + "?query=" + city + "&access_key=" + access_key + "&units=metric"

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
		Temperature:   float64(a.Current.Temperature),
		Feelslike:     float64(a.Current.FeelsLike),
		Pressure:      a.Current.Pressure,
		Humidity:      a.Current.Humidity,
		WindSpeed:     float64(a.Current.WindSpeed),
		WindDirection: a.Current.WindDegree,
		Visibility:    a.Current.Visibility * 1000, //convert to metres
	}, nil
}
