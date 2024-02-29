package main

import (
	"time"
)

type Station struct {
	Id        int
	StationId int `json:"station_id"`
	Name      string
	Lat       float64
	Lon       float64
	BikeCount int `json:"numBikesAvailable"`
	DockCount int `json:"numDocksAvailable"`
	Distance  int
	UpdateAt  time.Time
}
