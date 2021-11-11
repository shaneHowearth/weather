package openweathermap

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
			city:     "melbourne",
			getError: fmt.Errorf("fake response error"),
			outError: fmt.Errorf("fake response error"),
		},
		"io error": {
			city:         "melbourne",
			ioError:      fmt.Errorf("fake io error"),
			expectedResp: &http.Response{Body: fakeIORC, Status: "200 OK", StatusCode: http.StatusOK},
			outError:     fmt.Errorf("getWeather: reading response error fake io error"),
		},
		"upstream error": {
			city:         "melbourne",
			expectedResp: &http.Response{Body: fakeIORC, StatusCode: http.StatusBadRequest},
			outError:     fmt.Errorf("getWeather: got bad status 400"),
		},
		"json error": {
			city:         "melbourne",
			marshalError: fmt.Errorf("fake json error"),
			expectedResp: &http.Response{Body: fakeIORC, Status: "200 OK", StatusCode: http.StatusOK},
			outError:     fmt.Errorf("getWeather: unmarshalling response error fake response error"),
		},
		"melbourne": {
			city:         "melbourne",
			expectedResp: &http.Response{Body: fakeIORC, Status: "200 OK", StatusCode: http.StatusOK},
			expected: response{
				Temperature: float64(15.48),
				WindSpeed:   float64(2.68),
			},
			readResponse: []byte(`{
    "base": "stations",
    "clouds": {
        "all": 75
    },
    "cod": 200,
    "coord": {
        "lat": -37.814,
        "lon": 144.9633
    },
    "dt": 1636614234,
    "id": 2158177,
    "main": {
        "feels_like": 14.6,
        "humidity": 58,
        "pressure": 1001,
        "temp": 15.48,
        "temp_max": 16.79,
        "temp_min": 13.36
    },
    "name": "Melbourne",
    "sys": {
        "country": "AU",
        "id": 2008797,
        "sunrise": 1636571009,
        "sunset": 1636621495,
        "type": 2
    },
    "timezone": 39600,
    "visibility": 10000,
    "weather": [
        {
            "description": "broken clouds",
            "icon": "04d",
            "id": 803,
            "main": "Clouds"
        }
    ],
    "wind": {
        "deg": 139,
        "gust": 6.71,
        "speed": 2.68
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
			ow, err := NewOpenWeather("test app ID")
			assert.Nil(t, err)

			// Test
			output, err := ow.GetWeather(tc.city)

			if tc.outError == nil {
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, output)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
