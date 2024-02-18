package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

var Stations []Station

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

func handleHttpError(w http.ResponseWriter, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
	log.Print(err)
}

func main() {
	if os.Getenv("REFRESH_STATIONS") != "" {
		err := refreshStationInformation()
		if err != nil {
			panic(err)
		}
	}

	f, err := os.ReadFile("station_information.json")
	if errors.Is(err, os.ErrNotExist) {
		panic(errors.New("please re-run with REFRESH_STATIONS env var"))
	} else if err != nil {
		panic(err)
	}

	var rawStations struct {
		Data struct{ Stations []Station }
	}

	err = json.Unmarshal(f, &rawStations)
	if err != nil {
		panic(err)
	}

	Stations = rawStations.Data.Stations

	stationsController := StationsController{}
	indexController := IndexController{}

	http.HandleFunc("GET /{$}", indexController.Show)
	http.HandleFunc("GET /stations/closest", stationsController.ListClosest)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
