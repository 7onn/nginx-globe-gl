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

type IpDataGeoLocation struct {
	Latitude    float32 `json:"latitude"`
	Longitude   float32 `json:"longitude"`
	CityName    string  `json:"cityName"`
	CountryName string  `json:"countryName"`
	CountryCode string  `json:"countryCode"`
}

type GeoJsonProperty struct {
	Name      string  `json:"name"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	PopMax    int     `json:"pop_max"`
}

type GeoJsonFeature struct {
	Type       string          `json:"type"`
	Properties GeoJsonProperty `json:"properties"`
}

type GeoJson struct {
	Type     string           `json:"type"`
	Features []GeoJsonFeature `json:"features"`
}
