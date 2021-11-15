package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeProvider struct{}

var fakeResponse struct{ Temperature, WindSpeed float64 }
var fakeResponseErr error

func (f *fakeProvider) GetWeather(city string) (struct{ Temperature, WindSpeed float64 }, error) {
	return fakeResponse, fakeResponseErr
}

func TestWeatherHandler(t *testing.T) {
	base := "/v1/weather"
	testcases := map[string]struct {
		query  string
		status int
		body   struct {
			Temperature float64 `json:"temperature_degrees"`
			WindSpeed   float64 `json:"wind_speed"`
		}
		response struct{ Temperature, WindSpeed float64 }
		method   string // defaults to "GET"
		errBody  string
		myTime   time.Time // mandatory
		fakeErr  error
	}{
		"successful": {
			query:  "?city=melbourne",
			status: http.StatusOK,
			body: struct {
				Temperature float64 `json:"temperature_degrees"`
				WindSpeed   float64 `json:"wind_speed"`
			}{100, 150},
			response: struct{ Temperature, WindSpeed float64 }{100, 150},
			myTime:   time.Now(),
		},
		"wrong method": {
			query:   "?city=melbourne",
			status:  http.StatusMethodNotAllowed,
			method:  "POST",
			errBody: "Bad method - Go away!\n",
		},
		"no city": {
			status:  http.StatusBadRequest,
			errBody: "Bad Request\n",
		},
		"non-existant city": {
			query:   "?city=fake",
			status:  http.StatusBadRequest,
			errBody: "Sorry, don't know that city \"fake\"\n",
		},
		"too quick": {
			query:  "?city=melbourne",
			status: http.StatusOK,
			body: struct {
				Temperature float64 `json:"temperature_degrees"`
				WindSpeed   float64 `json:"wind_speed"`
			}{0, 0},
			myTime: time.Now().Add(-101 * 24 * 365 * time.Hour),
		},
		"failover": {
			query:  "?city=melbourne",
			status: http.StatusOK,
			body: struct {
				Temperature float64 `json:"temperature_degrees"`
				WindSpeed   float64 `json:"wind_speed"`
			}{0, 0},
			response: struct{ Temperature, WindSpeed float64 }{100, 150},
			myTime:   time.Now(),
			fakeErr:  fmt.Errorf("fake error"),
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			//create
			o, err := New([]Provider{&fakeProvider{}})
			assert.Nil(t, err)
			timeNow = func() time.Time { return tc.myTime }
			fakeResponse = tc.response
			fakeResponseErr = tc.fakeErr

			// Prepare request
			req, err := http.NewRequest(tc.method, base+tc.query, nil)
			assert.Nil(t, err)

			// make request
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o.Weather)
			handler.ServeHTTP(rr, req)

			// Check results
			assert.Equal(t, tc.status, rr.Code)

			if tc.errBody == "" {
				body := struct {
					Temperature float64 `json:"temperature_degrees"`
					WindSpeed   float64 `json:"wind_speed"`
				}{}
				fmt.Println(rr.Body.String())
				err = json.Unmarshal([]byte(rr.Body.String()), &body)
				assert.Nil(t, err)
				assert.Equal(t, tc.body, body)
			} else {
				assert.Equal(t, tc.errBody, rr.Body.String())
			}
		})
	}
}
