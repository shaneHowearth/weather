package openweathermap_test

import (
	"fmt"
	"testing"

	"github.com/shanehowearth/weather/providers/openweathermap"
	"github.com/stretchr/testify/assert"
)

func TestNewOpenWeather(t *testing.T) {

	testcases := map[string]struct {
		appID string
		err   error
	}{
		"successful creation": {
			appID: "test api key",
		},
		"no appID": {
			err: fmt.Errorf("appID is required"),
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ow, err := openweathermap.NewOpenWeather(tc.appID)
			if tc.err == nil {
				assert.Nil(t, err)
				assert.NotNil(t, ow)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
