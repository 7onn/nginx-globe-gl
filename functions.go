package main

import (
	"encoding/json"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/joho/godotenv/autoload"
	"github.com/oschwald/geoip2-golang"
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

func getCityFromIp(ip string, cities map[string]City) City {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal().Err(err).Msg("getIpDataGeoLocationFromIp - Failed to open GeoIP2-City.mmdb")
	}
	defer db.Close()

	parsedIp := net.ParseIP(ip)
	record, err := db.City(parsedIp)
	if err != nil {
		log.Fatal().Err(err).Msg("getIpDataGeoLocationFromIp - Failed to open GeoIP2-City.mmdb")
	}

	cityName := record.City.Names["en"]
	if cityName == "" {
		return City{}
	}

	// Prevent extremely large labels by setting the max evidence to 10
	var e int
	if cities[cityName].Evidence < 10 {
		e = cities[cityName].Evidence + 1
	} else {
		e = cities[cityName].Evidence
	}

	return City{
		CityName:  cityName,
		Latitude:  float32(record.Location.Latitude),
		Longitude: float32(record.Location.Longitude),
		Evidence:  e,
	}
}

func updateGeoLocations() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_HOST"),
		},
		Username: os.Getenv("ELASTICSEARCH_USER"),
		Password: os.Getenv("ELASTICSEARCH_PASSWORD"),
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("updateGeoLocations - Can not create Elasticsearch client")
	}

	res, err := es.Search(
		es.Search.WithBody(strings.NewReader(os.Getenv("ELASTICSEARCH_QUERY"))),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("updateGeoLocations - Can not query Elasticsearch")
	}

	defer res.Body.Close()

	var sr searchResult
	err = json.NewDecoder(res.Body).Decode(&sr)
	if err != nil {
		log.Fatal().Err(err).Msg("updateGeoLocations - Can not decode Elasticsearch response into searchResult struct")
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(sr.Hits.Hits))

	cities := map[string]City{}
	for _, h := range sr.Hits.Hits {
		go func(v hit) {
			ip := getIpFromLog(v.Source.Log)
			mu.Lock()
			city := getCityFromIp(ip, cities)
			if city.Latitude != 0 {
				cities[city.CityName] = city
			}
			mu.Unlock()
			wg.Done()
		}(h)
	}
	wg.Wait()

	geoJson := GeoJson{
		Type:     "FeatureCollection",
		Features: []GeoJsonFeature{},
	}

	for _, data := range cities {
		geoJsonFeature := GeoJsonFeature{
			Type: "Pointer",
			Properties: GeoJsonProperty{
				Latitude:  data.Latitude,
				Longitude: data.Longitude,
				PopMax:    data.Evidence * 500000,
				Name:      data.CityName,
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

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
