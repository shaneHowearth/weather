package weather

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Minimum time between requests
const minGap = 3 * time.Second

// Known cities
var cities = map[string]struct{}{"melbourne": struct{}{}}

// Provider -
// All weather providers should implement this interface to allow them to be
// used by the application.
type Provider interface {
	GetWeather(city string) (struct{ Temperature, WindSpeed float64 }, error)
}

type data struct {
	m         sync.Mutex
	providers []Provider
	last      struct {
		WindSpeed   float64 `json:"wind_speed"`
		Temperature float64 `json:"temperature_degrees"`
	}
	touched time.Time
}

// NewData -
// ignore linter warning on returning unexported type
// nolint:revive
func New(p []Provider) (*data, error) {
	// Must have at least one provider
	if len(p) < 1 {
		return nil, fmt.Errorf("must have at least one provider")
	}
	return &data{
		providers: p,
		touched:   time.Now().Add(-100 * 24 * 365 * time.Hour), // Default the last touched to 100 years ago
	}, nil
}

// Enable the following to be faked in tests
var timeNow = time.Now

// Weather -
func (d *data) Weather(w http.ResponseWriter, r *http.Request) {
	// only GET allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Bad method - Go away!", http.StatusMethodNotAllowed)
		return
	}

	// city is required
	query := r.URL.Query()
	city, ok := query["city"]
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// unknown city provided
	if _, ok := cities[strings.ToLower(strings.TrimSpace(city[0]))]; !ok {
		http.Error(w, "Sorry, don't know that city", http.StatusBadRequest)
		return
	}

	d.m.Lock()
	defer d.m.Unlock()
	// Rate limit
	// Note this limit is on this endpoint rather than provider specific
	if timeNow().Sub(d.touched) < minGap {
		// use the cached value
		if resp, err := json.Marshal(d.last); err == nil {
			_, err = w.Write(resp)
			if err != nil {
				log.Printf("unable to write cached %#v in GetWeather handler with error %v", d.last, err)
			}
		} else {
			log.Printf("unable to marshal cached %#v, with error %v", d.last, err)
		}
		return
	}

	// try each of the providers
	for i := range d.providers {
		val, err := d.providers[i].GetWeather(city[0])
		if err != nil {
			// log the error
			log.Printf("ERROR %v", err)
			continue
		}

		// Update cache
		d.touched = timeNow()
		d.last.Temperature = val.Temperature
		d.last.WindSpeed = val.WindSpeed

		// no need to try any more providers
		break
	}

	if resp, err := json.Marshal(d.last); err == nil {
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("unable to write %#v in GetWeather handler with error %v", d.last, err)
		}
	} else {
		log.Printf("unable to marshal %#v, with error %v", d.last, err)
	}
}
