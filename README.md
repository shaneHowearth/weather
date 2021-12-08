# Weather integration example

A simple RESTful API that consumes from upstream Weather Services and produces
temperature and wind speed information for the desired city.

# Build and Run

Environment variables need to be set for this example:
* HTTP_PORT - the port that the service will listen on
* OPENWEATHER - api key for openweathermap.org
* WEATHERSTACK - api key for weatherstack.com

Then use the command `docker compose up` or `go run cmd/main.go` to run the
service.

Using a tool like curl you can interact with the applications API
eg. The following example `curl localhost:8080/v1/weather?city=melbourne`
will return a json object similar to this:
`{"wind_speed":2.68,"temperature_degrees":12.26}`,

If an unknown city is provided an error message (Sorry, don't know that city)
will be returned, and the status will be 400.

# Limitations
Currently the application only supports lookup for Melbourne and Sydney (both
Australia), adding more
supported locations involves updating a central store, and a store for each
provider that translates the location name provided to something that it
understands.
More providers can be added by implementing the weather.Provider interface, and
injecting an instance of that provider into the weather.data (done in
cmd/main.go when the weather.data is instantiated).

# Unit tests
All tests can be run with `go test ./...`
