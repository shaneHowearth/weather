package weather_test

import (
	"fmt"
	"testing"

	"github.com/shanehowearth/weather"
	"github.com/stretchr/testify/assert"
)

type fakeProvider struct{}

func (f *fakeProvider) GetWeather(city string) (struct{ Temperature, WindSpeed float64 }, error) {
	return struct{ Temperature, WindSpeed float64 }{}, nil
}
func TestNew(t *testing.T) {
	testcases := map[string]struct {
		providers []weather.Provider
		err       error
	}{
		"no providers": {
			err: fmt.Errorf("must have at least one provider"),
		},
		"succesful creation": {
			providers: []weather.Provider{&fakeProvider{}},
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			o, err := weather.New(tc.providers)
			if tc.err == nil {
				assert.Nil(t, err)
				assert.NotNil(t, o)
			} else {
				assert.NotNil(t, err)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}
