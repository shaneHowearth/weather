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
var cities = map[string]struct{}{"melbourne": struct{}{}, "sydney": struct{}{}}

// Provider -
// All weather providers should implement this interface to allow them to be
// used by the application.
type Provider interface {
	GetWeather(city string) (struct{ Temperature, WindSpeed float64 }, error)
}

type data struct {
	m         sync.Mutex
	providers []Provider
	last      map[string]struct {
		WindSpeed   float64 `json:"wind_speed"`
		Temperature float64 `json:"temperature_degrees"`
	}
	touched map[string]time.Time
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
		touched:   map[string]time.Time{},
		last: map[string]struct {
			WindSpeed   float64 `json:"wind_speed"`
			Temperature float64 `json:"temperature_degrees"`
		}{},
	}, nil
}

// Enable the following to be faked in tests
var timeNow = time.Now

// Weather -
func (d *data) Weather(w http.ResponseWriter, r *http.Request) {
	// only GET allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Bad method", http.StatusMethodNotAllowed)
		return
	}

	// city is required
	query := r.URL.Query()
	cityQuery, ok := query["city"]
	if !ok {
		http.Error(w, "Bad Request, unknown city", http.StatusBadRequest)
		return
	}
	city := cityQuery[0]

	// unknown city provided
	if _, ok := cities[strings.ToLower(strings.TrimSpace(city))]; !ok {
		e := fmt.Sprintf("Sorry, don't know that city %q", city)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	d.m.Lock()
	defer d.m.Unlock()
	// Rate limit
	// Note this limit is on this endpoint rather than specific provider
	if timeNow().Sub(d.touched[city]) < minGap {
		// use the cached value
		if resp, err := json.Marshal(d.last[city]); err == nil {
			_, err = w.Write(resp)
			if err != nil {
				log.Printf("unable to write cached %#v in GetWeather handler with error %v", d.last[city], err)
			}
		} else {
			log.Printf("unable to marshal cached %#v, with error %v", d.last[city], err)
		}
		return
	}

	// try each of the providers
	for i := range d.providers {
		val, err := d.providers[i].GetWeather(city)
		if err != nil {
			// log the error
			log.Printf("ERROR %v", err)
			continue
		}

		// Update cache
		d.touched[city] = timeNow()
		d.last[city] = struct {
			WindSpeed   float64 `json:"wind_speed"`
			Temperature float64 `json:"temperature_degrees"`
		}{
			Temperature: val.Temperature,
			WindSpeed:   val.WindSpeed,
		}

		// no need to try any more providers
		break
	}

	if resp, err := json.Marshal(d.last[city]); err == nil {
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("unable to write %#v in GetWeather handler with error %v", d.last[city], err)
		}
	} else {
		log.Printf("unable to marshal %#v, with error %v", d.last[city], err)
	}
}
