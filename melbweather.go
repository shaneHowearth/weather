package melbweather

// WeatherProvider -
// All weather providers should implement this interface to allow them to be
// used by the application.
type WeatherProvider interface {
	GetWeather(city string) (struct{ Temperature, WindSpeed float64 }, error)
	GetCity(string), string
}

