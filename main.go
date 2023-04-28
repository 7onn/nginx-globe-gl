package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
)

func main() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"",
		},
		Username: "",
		Password: "",
	}
	es, _ := elasticsearch.NewClient(cfg)
	res, _ := es.Search(
		es.Search.WithBody(strings.NewReader(`
		{
			"from" : 0, 
			"size" : 1000,
			"query": {
				"match" : {
					"kubernetes.labels.app_kubernetes_io/name": "ingress-nginx"
				}
			},
			"sort" : [
	  		{ "@timestamp" : "desc" }
			]
		}
	  `)),

		es.Search.WithPretty(),
	)
	defer res.Body.Close()

	var sr searchResult
	err := json.NewDecoder(res.Body).Decode(&sr)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	http.HandleFunc("/locations.geojson", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "locations.geojson")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		<html>
		<head>
			<style> body { margin: 0; } </style>
			<script src="//unpkg.com/globe.gl"></script>
			<!--<script src="../../dist/globe.gl.js"></script>-->
		</head>

		<body>
			<div id="globeViz"></div>
			<script>
				fetch('%s').then(res => res.json()).then(places => {
					const world = Globe()
						.globeImageUrl('//unpkg.com/three-globe/example/img/earth-night.jpg')
						.backgroundImageUrl('//unpkg.com/three-globe/example/img/night-sky.png')
						.labelsData(places.features)
						.labelLat(d => d.properties.latitude)
						.labelLng(d => d.properties.longitude)
						.labelText(d => d.properties.name)
						.labelSize(d => Math.sqrt(d.properties.pop_max) * 4e-4)
						.labelDotRadius(d => Math.sqrt(d.properties.pop_max) * 4e-4)
						.labelColor(() => 'rgba(255, 165, 0, 0.75)')
						.labelResolution(2)
					(document.getElementById('globeViz'))

					world.controls().autoRotate = true;
					world.controls().autoRotateSpeed = 3;
				});
			</script>
			</body>
		</html>
	`, "http://localhost:9999/locations.geojson")
	})

	fmt.Println("Server listening on port 9999")
	http.ListenAndServe(":9999", nil)
}
