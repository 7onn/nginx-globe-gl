package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

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

	globeGLFileSchema := GlobeGLFileSchema{}

	// geoJson := GeoJson{
	// 	Type:     "FeatureCollection",
	// 	Features: []GeoJsonFeature{},
	// }

	for _, v := range sr.Hits.Hits {
		ip := getIpFromLog(v.Source.Log)
		geolocation := geoLocationFromIp(ip)

		if geolocation.Latitude != 0 {
			globeGLFileSchema.GeoLocations = append(globeGLFileSchema.GeoLocations, geolocation)

			// geoJson = append(geoJson.Features, geoJsonFeature)
		}

	}

	fmt.Println(globeGLFileSchema)

	file, err := os.Create("geolocations.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(globeGLFileSchema)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
			<html>
			<head>
				<style> body { margin: 0; } </style>
				<script src="//unpkg.com/globe.gl"></script>
				<!--<script src="../../dist/globe.gl.js"></script>-->
			</head>

			<body>
				<div id="globeViz"></div>
				<script>
					fetch('/Users/tom/l/nginx-globe-gl/geolocations.geojson').then(res => res.json()).then(places => {
						Globe()
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
					});
				</script>
				</body>
			</html>
		`

		fmt.Fprintf(w, html)
	})

	fmt.Println("Server listening on port 9999")
	http.ListenAndServe(":9999", nil)
}
