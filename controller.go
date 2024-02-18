package main

type Station struct {
	StationCode       string
	NumBikesAvailable int
	NumDocksAvailable int
	Name              string
	Lat               float64
	Lon               float64
	Distance          int
}

type Controller struct {
	Stations []Station
}
