package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

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
