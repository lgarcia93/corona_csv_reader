package models

import "time"

// Region ....
type Region struct {
	Key             string  `json:"key"`
	Province        string  `json:"province"`
	Country         string  `json:"country"`
	LastUpdated     string  `json:"lastUpdated"`
	ConfirmedCases  int     `json:"confirmedCases"`
	ConfirmedDeaths int     `json:"confirmedDeaths"`
	Recovered       int     `json:"recovered"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

// GeocodingResponse ...
type GeocodingResponse struct {
	Results []Result `json:"results"`
}

// Result ...
type Result struct {
	Coords LatLng `json:"geometry"`
}

// LatLng ...
type LatLng struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// TimeSeriesItem ...
type TimeSeriesItem struct {
	Date            time.Time `json:"date"`
	ConfirmedCases  int       `json:"confirmedCases"`
	ConfirmedDeaths int       `json:"confirmedDeaths"`
	Recoveries      int       `json:"recoveries"`
}
