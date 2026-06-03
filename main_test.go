package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type FullResponse struct {
	DC      DataCenter
	Weather WeatherData
}

func TestPopulateGoogleAPIData(t *testing.T) {
	const workers = 4
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}

	DataCenterChannel := make(chan DataCenter, len(DataCenters))
	go func() {
		for _, dc := range DataCenters {
			DataCenterChannel <- dc
		}
		close(DataCenterChannel)
	}()

	Responses := make(chan FullResponse, len(DataCenters))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dc := range DataCenterChannel {
				weatherData, err := PopulateGoogleAPIData(dc)
				if err != nil {
					t.Errorf("Error populating Google APIData: %v", err)
				}
				select {
				case Responses <- FullResponse{dc, weatherData}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(Responses)
	}()

	for {
		select {
		case response, ok := <-Responses:
			if !ok {
				return
			}
			fmt.Printf("The weather in %s is %f degrees C\n", response.DC.Address, response.Weather.Temperature.Degrees)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}

}
