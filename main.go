package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"slices"
)

type RawStations struct {
	Data struct{ Stations []Station }
}

type Station struct {
	StationCode       string
	NumBikesAvailable int
	Name              string
	Lat               float64
	Lon               float64
	Distance          float64
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	lat1 = lat1 * math.Pi / 180
	lon1 = lon1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	lon2 = lon2 * math.Pi / 180

	R := 6371.0

	dlon := lon2 - lon1
	dlat := lat2 - lat1

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
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

func main() {

	shouldRefreshStations := os.Getenv("REFRESH_STATIONS") != ""

	if shouldRefreshStations {
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

	stations := make(map[string]Station)

	// // me
	// myLat := 48.8372335
	// myLon := 2.3890275

	// adi
	myLat := 48.844989
	myLon := 2.309301

	for _, station := range rawStations.Data.Stations {
		station.Distance = haversine(myLat, myLon, station.Lat, station.Lon)
		stations[station.StationCode] = station
	}

	r, err := http.Get("https://velib-metropole-opendata.smovengo.cloud/opendata/Velib_Metropole/station_status.json")
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&rawStations)

	for _, station := range rawStations.Data.Stations {
		if s, ok := stations[station.StationCode]; ok {
			s.NumBikesAvailable = station.NumBikesAvailable
			stations[station.StationCode] = s
		}
	}

	var closestStations []Station
	for _, station := range stations {
		if station.Distance < 1 {
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

	for _, station := range closestStations {
		fmt.Println(int(station.Distance*1000), "m", station.NumBikesAvailable, station.Name)
	}
}
