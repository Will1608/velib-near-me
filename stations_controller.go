package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
)

type StationsController struct{}

func (s StationsController) ListClosest(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT name, lat, lon, dock_count, bike_count FROM stations WHERE dock_count > 0 OR bike_count > 0")
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	var stations []Station
	for rows.Next() {
		var station Station
		err := rows.Scan(&station.Name, &station.Lat, &station.Lon, &station.DockCount, &station.BikeCount)
		if err != nil {
			defer handleHttpError(w, err)
			return
		}

		stations = append(stations, station)
	}

	params := r.URL.Query()
	latitude, err := strconv.ParseFloat(params.Get("latitude"), 64)
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	longitude, err := strconv.ParseFloat(params.Get("longitude"), 64)
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	for i, station := range stations {
		stations[i].Distance = Haversine(latitude, longitude, station.Lat, station.Lon)
	}
	slices.SortFunc(stations, func(a Station, b Station) int { return a.Distance - b.Distance })

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(stations[:5])
	if err != nil {
		defer handleHttpError(w, err)
		return
	}
}
