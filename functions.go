package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
)

func verifyEnv() {
	esHost := os.Getenv("ELASTICSEARCH_HOST")
	if esHost == "" {
		log.Fatal().Msg("verifyEnv - Can not find environment variable ELASTICSEARCH_HOST is not set")
	}

	esUser := os.Getenv("ELASTICSEARCH_USER")
	if esUser == "" {
		log.Fatal().Msg("verifyEnv - Can not find environment variable ELASTICSEARCH_USER is not set")
	}

	esPassword := os.Getenv("ELASTICSEARCH_PASSWORD")
	if esPassword == "" {
		log.Fatal().Msg("verifyEnv - Can not find environment variable ELASTICSEARCH_PASSWORD is not set")
	}

	selfURL := os.Getenv("SELF_URL")
	if selfURL == "" {
		log.Fatal().Msg("verifyEnv - Can not find environment variable SELF_URL is not set")
	}

	esQuery := os.Getenv("ELASTICSEARCH_QUERY")
	if esQuery == "" {
		log.Fatal().Msg("verifyEnv - Can not find environment variable ELASTICSEARCH_QUERY is not set")
	}
}

func getIpFromLog(log string) string {
	re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`) // Compile a regular expression to match words
	return re.FindString(log)
}

func getIpDataGeoLocationFromIp(ip string) IpDataGeoLocation {
	values := url.Values{}
	values.Set("ip", ip)
	values.Set("source", "ip2location")
	values.Set("ipv", "4")

	req, err := http.NewRequest("POST", "https://www.iplocation.net/get-ipdata", strings.NewReader(values.Encode()))
	if err != nil {
		log.Fatal().Err(err).Msg("getIpDataGeoLocationFromIp - Can not create request client")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("getIpDataGeoLocationFromIp - Can not request iplocation.net")
		return IpDataGeoLocation{}
	}
	defer resp.Body.Close()

	var res ipData
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		log.Error().Err(err).Msg("getIpDataGeoLocationFromIp - Can not decode iplocation.net response into ipData struct")
		return IpDataGeoLocation{}
	}

	return res.GeoLocation
}

func updateGeoLocations() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_HOST"),
		},
		Username: os.Getenv("ELASTICSEARCH_USER"),
		Password: os.Getenv("ELASTICSEARCH_PASSWORD"),
	}
	es, _ := elasticsearch.NewClient(cfg)
	res, _ := es.Search(
		es.Search.WithBody(strings.NewReader(os.Getenv("ELASTICSEARCH_QUERY"))),
	)
	defer res.Body.Close()

	var sr searchResult
	err := json.NewDecoder(res.Body).Decode(&sr)
	if err != nil {
		log.Fatal().Err(err).Msg("updateGeoLocations - Can not decode Elasticsearch response into searchResult struct")
	}

	geoLocationEvidence := map[IpDataGeoLocation]int{}

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(sr.Hits.Hits))

	for _, h := range sr.Hits.Hits {
		go func(v hit) {
			ip := getIpFromLog(v.Source.Log)
			geolocation := getIpDataGeoLocationFromIp(ip)
			if geolocation.Latitude != 0 {
				mu.Lock()
				geoLocationEvidence[geolocation]++
				mu.Unlock()
			}
			wg.Done()
		}(h)
	}
	wg.Wait()

	geoJson := GeoJson{
		Type:     "FeatureCollection",
		Features: []GeoJsonFeature{},
	}

	for gl, qty := range geoLocationEvidence {
		geoJsonFeature := GeoJsonFeature{
			Type: "Pointer",
			Properties: GeoJsonProperty{
				Latitude:  gl.Latitude,
				Longitude: gl.Longitude,
				PopMax:    qty * 1000000,
				Name:      gl.CityName,
			},
		}
		geoJson.Features = append(geoJson.Features, geoJsonFeature)
	}

	file, err := os.Create("locations.geojson")
	if err != nil {
		log.Fatal().Err(err).Msg("updateGeoLocations - Can not create locations.geojson file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(geoJson)
	if err != nil {
		log.Fatal().Err(err).Msg("updateGeoLocations - Can not write into locations.geojson file")
	}
}
