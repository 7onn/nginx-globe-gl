package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/joho/godotenv/autoload"
)

func verifyEnv() {
	esHost := os.Getenv("ELASTICSEARCH_HOST")
	if esHost == "" {
		log.Fatal("Environment variable ELASTICSEARCH_HOST is not set")
	}

	esUser := os.Getenv("ELASTICSEARCH_USER")
	if esUser == "" {
		log.Fatal("Environment variable ELASTICSEARCH_USER is not set")
	}

	esPassword := os.Getenv("ELASTICSEARCH_PASSWORD")
	if esPassword == "" {
		log.Fatal("Environment variable ELASTICSEARCH_PASSWORD is not set")
	}

	selfURL := os.Getenv("SELF_URL")
	if selfURL == "" {
		log.Fatal("Environment variable SELF_URL is not set")
	}

	esQuery := os.Getenv("ELASTICSEARCH_QUERY")
	if esQuery == "" {
		log.Fatal("Environment variable ELASTICSEARCH_QUERY is not set")
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
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// io.Copy(os.Stdout, resp.Body)

	var res ipData
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		panic(err)
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
		fmt.Println(err)
		time.Sleep(5 * time.Second)
		return
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
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(geoJson)
	if err != nil {
		fmt.Println(err)
		time.Sleep(5 * time.Second)
		return
	}
}
