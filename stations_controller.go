package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"slices"
	"strconv"
)

type StationsController struct{}

func (s StationsController) ListClosest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var queryLat, queryLon float64

	if query.Get("latitude") != "" {
		lat, err := strconv.ParseFloat(query.Get("latitude"), 64)
		if err != nil {
			defer handleHttpError(w, err)
			return
		}
		queryLat = lat
	}

	if query.Get("longitude") != "" {
		lon, err := strconv.ParseFloat(query.Get("longitude"), 64)
		if err != nil {
			defer handleHttpError(w, err)
			return
		}
		queryLon = lon
	}

	stationsMap := make(map[string]Station)
	for _, station := range Stations {
		station.Distance = Haversine(queryLat, queryLon, station.Lat, station.Lon)
		stationsMap[station.StationCode] = station
	}

	req, err := http.Get("https://velib-metropole-opendata.smovengo.cloud/opendata/Velib_Metropole/station_status.json")
	if err != nil {
		defer handleHttpError(w, err)
		return
	}
	defer req.Body.Close()

	var stationStatus struct{ Data struct{ Stations []Station } }
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&stationStatus)

	for _, station := range stationStatus.Data.Stations {
		if s, ok := stationsMap[station.StationCode]; ok {
			s.NumBikesAvailable = station.NumBikesAvailable
			s.NumDocksAvailable = station.NumDocksAvailable
			stationsMap[station.StationCode] = s
		}
	}

	var closestStations []Station
	for _, station := range stationsMap {
		if station.Distance < 500 {
			closestStations = append(closestStations, station)
		}
	}

	slices.SortFunc(closestStations, func(a Station, b Station) int {
		if a.Distance >= b.Distance {
			return 1
		} else {
			return -1
		}
	})

	tmpl, err := template.ParseFiles("closest-stations.html")
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	var returning bool
	if query.Get("action") == "returning" {
		returning = true
	}

	err = tmpl.Execute(w, struct {
		Stations  []Station
		Returning bool
	}{
		Stations:  closestStations,
		Returning: returning,
	})
	if err != nil {
		handleHttpError(w, err)
	}
}
