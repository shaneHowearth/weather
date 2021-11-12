package weatherstack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeIOReadCloser struct{}

var readResponse []byte

func (f *fakeIOReadCloser) Read(p []byte) (n int, err error) {
	return len(readResponse), nil
}
func (f *fakeIOReadCloser) Close() error {
	return nil
}

func TestGetWeather(t *testing.T) {
	fakeIORC := &fakeIOReadCloser{}

	testcases := map[string]struct {
		city         string
		getError     error
		ioError      error
		marshalError error
		outError     error
		expectedResp *http.Response
		expected     response
		readResponse []byte
	}{
		"no city": {
			outError: fmt.Errorf("city is required"),
		},
		"http error": {
			city:     "Melbourne",
			getError: fmt.Errorf("fake response error"),
			outError: fmt.Errorf("fake response error"),
		},
		"io error": {
			city:         "Melbourne",
			ioError:      fmt.Errorf("fake io error"),
			expectedResp: &http.Response{Body: fakeIORC, Status: "200 OK", StatusCode: http.StatusOK},
			outError:     fmt.Errorf("getWeather: reading response error fake io error"),
		},
		"upstream error": {
			city:         "Melbourne",
			expectedResp: &http.Response{Body: fakeIORC, StatusCode: http.StatusBadRequest},
			outError:     fmt.Errorf("getWeather: got bad status 400"),
		},
		"json error": {
			city:         "Melbourne",
			marshalError: fmt.Errorf("fake json error"),
			expectedResp: &http.Response{Body: fakeIORC, Status: "200 OK", StatusCode: http.StatusOK},
			outError:     fmt.Errorf("getWeather: unmarshalling response error fake response error"),
		},
		"melbourne": {
			city:         "Melbourne",
			expectedResp: &http.Response{Body: fakeIORC, Status: "200 OK", StatusCode: http.StatusOK},
			expected: response{
				Temperature: float64(15),
				WindSpeed:   float64(28),
			},
			readResponse: []byte(`{
    "current": {
        "cloudcover": 0,
        "feelslike": 14,
        "humidity": 55,
        "is_day": "yes",
        "observation_time": "06:59 AM",
        "precip": 0,
        "pressure": 1004,
        "temperature": 15,
        "uv_index": 5,
        "visibility": 10,
        "weather_code": 113,
        "weather_descriptions": [
            "Sunny"
        ],
        "weather_icons": [
            "https://assets.weatherstack.com/images/wsymbols01_png_64/wsymbol_0001_sunny.png"
        ],
        "wind_degree": 170,
        "wind_dir": "S",
        "wind_speed": 28
    },
    "location": {
        "country": "Australia",
        "lat": "-37.817",
        "localtime": "2021-11-11 17:59",
        "localtime_epoch": 1636653540,
        "lon": "144.967",
        "name": "Melbourne",
        "region": "Victoria",
        "timezone_id": "Australia/Melbourne",
        "utc_offset": "11.0"
    },
    "request": {
        "language": "en",
        "query": "Melbourne, Australia",
        "type": "City",
        "unit": "m"
    }
}`),
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// Set up
			httpGet = func(url string) (resp *http.Response, err error) {
				return tc.expectedResp, tc.getError
			}

			ioutilReadAll = func(r io.Reader) ([]byte, error) {
				return tc.readResponse, tc.ioError
			}
			jsonUnmarshal = func(data []byte, v interface{}) error {
				if tc.marshalError == nil {
					return json.Unmarshal(data, v)
				}
				return tc.marshalError
			}
			ws, err := NewWeatherStack("Test Access Key")
			assert.Nil(t, err)

			// Test
			output, err := ws.GetWeather(tc.city)

			if tc.outError == nil {
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, output)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
