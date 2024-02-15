package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
)

type RawStations struct {
	Data struct{ Stations []Station }
}

type Station struct {
	StationCode       string
	NumBikesAvailable int
	NumDocksAvailable int
	Name              string
	Lat               float64
	Lon               float64
	Distance          int
}

func haversine(lat1, lon1, lat2, lon2 float64) int {
	lat1 = lat1 * math.Pi / 180
	lon1 = lon1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	lon2 = lon2 * math.Pi / 180

	R := 6371.0

	dlon := lon2 - lon1
	dlat := lat2 - lat1

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int(1000 * R * c)
}

func refreshStationInformation() error {
	r, err := http.Get("https://velib-metropole-opendata.smovengo.cloud/opendata/Velib_Metropole/station_information.json")
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	stationInformation, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	return os.WriteFile("station_information.json", stationInformation, 0600)
}

func nearestStations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var queryLat, queryLon float64
	if v, ok := query["latitude"]; ok && len(v) != 0 {
		lat, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			panic(err)
		}
		queryLat = lat
	}

	if v, ok := query["longitude"]; ok && len(v) != 0 {
		lon, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			panic(err)
		}
		queryLon = lon
	}

	stationsMap := make(map[string]Station)
	for _, station := range slices.Clone(stations) {
		station.Distance = haversine(queryLat, queryLon, station.Lat, station.Lon)
		stationsMap[station.StationCode] = station
	}

	req, err := http.Get("https://velib-metropole-opendata.smovengo.cloud/opendata/Velib_Metropole/station_status.json")
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	var stationStatus RawStations
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

	// query := r.URL.Query()
	// var returning bool
	// if v, ok := query["action"]; ok && len(v) != 0 {
	// 	returning = v[0] == "retruning"
	// }

	tmpl, err := template.ParseFiles("closest-stations.html")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, struct {
		Stations []Station
	}{
		Stations: closestStations,
	})
	if err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(w, nil)
}

var stations []Station

func main() {
	if os.Getenv("REFRESH_STATIONS") != "" {
		err := refreshStationInformation()
		if err != nil {
			panic(err)
		}
	}

	f, err := os.ReadFile("station_information.json")
	if errors.Is(err, os.ErrNotExist) {
		err := refreshStationInformation()
		if err != nil {
			panic(err)
		}
		f, err = os.ReadFile("station_information.json")
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	var rawStations RawStations
	err = json.Unmarshal(f, &rawStations)
	if err != nil {
		panic(err)
	}
	stations = rawStations.Data.Stations

	http.HandleFunc("GET /{$}", index)
	http.HandleFunc("GET /stations/closest", nearestStations)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
