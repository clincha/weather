package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestPopulateGoogleAPIData(t *testing.T) {

	WeatherChan := make(chan WeatherData, 4)

	DataCenters := [4]DataCenter{
		{
			Address: "London",
			GeoCode: GeoCode{},
		},
		{
			Address: "Bristol",
			GeoCode: GeoCode{},
		},
		{
			Address: "Edinburgh",
			GeoCode: GeoCode{},
		},
		{
			Address: "Melbourne",
			GeoCode: GeoCode{},
		},
	}

	wg := sync.WaitGroup{}

	wg.Add(len(DataCenters))
	go func() {
		wg.Wait()
		close(WeatherChan)
	}()
	for _, dc := range DataCenters {
		go func(dc DataCenter) {
			defer wg.Done()
			weatherData, err := PopulateGoogleAPIData(dc)
			if err != nil {
				t.Errorf("Error populating Google APIData: %v", err)
			}

			if weatherData.Temperature.Unit == "" {
				t.Errorf("Temperature unit not populated")
			}

			if weatherData.Temperature.Degrees == 0 {
				t.Errorf("Temperature degrees not populated")
			}
			WeatherChan <- weatherData
		}(dc)
	}

	for weather := range WeatherChan {
		fmt.Println(weather)
	}

}
