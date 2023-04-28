package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	verifyEnv()

	go func() {
		for {
			updateGeoLocations()
			time.Sleep(time.Second * 60)
		}
	}()

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
			  fetch('%s/locations.geojson').then(res => res.json()).then(places => {
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

			  	world.controls().autoRotate = true
			  	world.controls().autoRotateSpeed = 3
			  });				
			</script>
		</body>
		</html>
	`, os.Getenv("SELF_URL"))
	})

	fmt.Println("Server listening on port 9999")
	log.Fatal(http.ListenAndServe(":9999", nil))
}
