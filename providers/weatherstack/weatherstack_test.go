package weatherstack_test

import (
	"fmt"
	"testing"

	"github.com/shanehowearth/weather/providers/weatherstack"
	"github.com/stretchr/testify/assert"
)

func TestNewWeatherStack(t *testing.T) {

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
			ws, err := weatherstack.NewWeatherStack(tc.appID)
			if tc.err == nil {
				assert.Nil(t, err)
				assert.NotNil(t, ws)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
