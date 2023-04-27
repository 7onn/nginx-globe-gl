package main

type source struct {
	Log string `json:"log"`
}

type hit struct {
	Source source `json:"_source"`
}

type hits struct {
	Hits []hit `json:"hits"`
}
type searchResult struct {
	Hits hits `json:"hits"`
}

type ipData struct {
	GeoLocation geoLocation `json:"res"`
}

type geoLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CityName    string  `json:"cityName"`
	CountryName string  `json:"countryName"`
	CountryCode string  `json:"countryCode"`
}

type GlobeGLFileSchema struct {
	GeoLocations []geoLocation
}

type GeoJsonProperty struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GeoJsonGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type GeoJsonFeature struct {
	Type       string          `json:"type"`
	Properties GeoJsonProperty `json:"properties"`
	Geometry   GeoJsonGeometry `json:"geometry"`
}

type GeoJson struct {
	Type     string           `json:"type"`
	Features []GeoJsonFeature `json:"features"`
}
