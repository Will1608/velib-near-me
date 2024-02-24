package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"slices"
	"strconv"
)

type StationsController struct{}

func (s StationsController) ListClosest(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	action := params.Get("action")

	var rows *sql.Rows
	var err error
	if action == "returning" {
		rows, err = db.Query("SELECT name, lat, lon, dock_count FROM stations WHERE dock_count > 0")
	} else {
		rows, err = db.Query("SELECT name, lat, lon, bike_count FROM stations WHERE bike_count > 0")
	}
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	var stations []Station
	for rows.Next() {
		var station Station
		var err error
		if action == "returning" {
			err = rows.Scan(&station.Name, &station.Lat, &station.Lon, &station.DockCount)
		} else {
			err = rows.Scan(&station.Name, &station.Lat, &station.Lon, &station.BikeCount)
		}
		if err != nil {
			defer handleHttpError(w, err)
			return
		}

		stations = append(stations, station)
	}

	lat, err := strconv.ParseFloat(params.Get("lat"), 64)
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	lon, err := strconv.ParseFloat(params.Get("lon"), 64)
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	for i, station := range stations {
		stations[i].Distance = Haversine(lat, lon, station.Lat, station.Lon)
	}
	slices.SortFunc(stations, func(a Station, b Station) int { return a.Distance - b.Distance })

	tmpl, err := template.ParseFiles("closest-stations.html")
	if err != nil {
		defer handleHttpError(w, err)
		return
	}

	tmpl.Execute(w, struct{ Stations []Station }{Stations: stations[:5]})
	if err != nil {
		defer handleHttpError(w, err)
		return
	}
}
