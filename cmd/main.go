package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/shanehowearth/weather"
	"github.com/shanehowearth/weather/providers/openweathermap"
	"github.com/shanehowearth/weather/providers/weatherstack"
)

func main() {
	rPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Fatal("HTTP_PORT required for HTTP server to listen on")
	}

	// Port validation - has to be converted to an int for the checks
	port, err := strconv.Atoi(rPort)
	if err != nil {
		log.Fatal("HTTP_PORT must be an integer")
	}
	// PORT must be non-privileged and legit
	if port <= 1024 || port >= 65535 {
		log.Fatal("HTTP_PORT must be between 1024 and 65535 (exclusive)")
	}

	// Weather providers

	// Open Weather App ID
	owID, ok := os.LookupEnv("OPENWEATHER")
	if !ok {
		log.Fatal("OPENWEATHER app id required")
	}
	ow, err := openweathermap.NewOpenWeather(owID)
	if err != nil {
		log.Fatalf("Unable to create new openweathermap provider instance, with error: %v", err)
	}

	// Weather stack Access Key
	wsKey, ok := os.LookupEnv("WEATHERSTACK")
	if !ok {
		log.Fatal("WEATHERSTACK access key required")
	}
	ws, err := weatherstack.NewWeatherStack(wsKey)
	if err != nil {
		log.Fatalf("Unable to create new weatherstack provider instance, with error: %v", err)
	}

	//
	w, err := weather.New([]weather.Provider{ow, ws})
	if err != nil {
		log.Fatalf("Unable to create new weather instance, with error: %v", err)
	}

	mux := http.NewServeMux()
	// Routes - note, in a more complex application routes would go into a
	// dedicated file
	mux.Handle("/v1/weather", http.HandlerFunc(w.Weather))

	// listen on all localhost
	ip := "127.0.0.1"
	server := &http.Server{Addr: ip + ":" + rPort, Handler: mux}

	// Server listens on its own goroutine
	go func() {
		log.Printf("Listening on %s:%s...", ip, rPort)
		if err := server.ListenAndServe(); err != nil {
			log.Panicf("Listen and serve returned error: %v", err)
		}
	}()

	// Graceful shutdown!

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown returned error %v", err)
	}
}
