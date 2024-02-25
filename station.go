package main

import (
	"math/big"
	"time"
)

type Station struct {
	Id        int
	StationId *big.Int `json:"station_id"`
	Name      string
	Lat       float64
	Lon       float64
	BikeCount int `json:"numBikesAvailable"`
	DockCount int `json:"numDocksAvailable"`
	Distance  int
	UpdateAt  time.Time
}
