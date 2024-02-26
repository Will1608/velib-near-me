package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func handleHttpError(w http.ResponseWriter, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
	log.Print(err)
}

var db *sql.DB

func refreshStations() error {
	var data struct {
		Data struct {
			Stations []Station
		}
	}

	// station information
	r, err := http.Get("https://velib-metropole-opendata.smovengo.cloud/opendata/Velib_Metropole/station_information.json")
	if err != nil {
		return err
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}

	// station status
	r, err = http.Get("https://velib-metropole-opendata.smovengo.cloud/opendata/Velib_Metropole/station_status.json")
	if err != nil {
		return err
	}
	defer r.Body.Close()

	dec = json.NewDecoder(r.Body)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})

	insertQuery := "INSERT INTO stations (station_id, name, lat, lon, bike_count, dock_count, updated_at) VALUES "
	for _, station := range data.Data.Stations {
		insertQuery += fmt.Sprintf("(%d, '%s', %f, %f, %d, %d, NOW()),", station.StationId, strings.Replace(station.Name, "'", "''", -1), station.Lat, station.Lon, station.BikeCount, station.DockCount)
	}
	insertQuery = strings.TrimRight(insertQuery, ",") + " ON CONFLICT (station_id) DO UPDATE SET bike_count = EXCLUDED.bike_count, dock_count = EXCLUDED.dock_count, updated_at = EXCLUDED.updated_at"
	_, err = tx.Exec(insertQuery)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	_, err = tx.Exec("DELETE FROM stations WHERE updated_at - NOW() > INTERVAL '1 minute'")
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}

func main() {
	var err error
	db, err = sql.Open("postgres", "postgresql://postgres@/velib?host=/var/run/postgresql/")
	if err != nil {
		panic(err)
	}

	// update data periodically
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				err := refreshStations()
				if err != nil {
					log.Print(err)
				}
			}
		}
	}()

	stationsController := StationsController{}
	indexController := IndexController{}

	http.HandleFunc("GET /{$}", indexController.Show)
	http.HandleFunc("GET /stations/closest", stationsController.ListClosest)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
