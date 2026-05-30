package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var GoogleApiKey = os.Getenv("GOOGLE_API_KEY")

type DataCenter struct {
	Address string
	GeoCode GeoCode
	Weather WeatherData
}

type GeocodeResponse struct {
	Results []GeoCode `json:"results"`
}

type GeoCode struct {
	PlaceID  string `json:"placeId"`
	Location LatLng `json:"location"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type WeatherData struct {
	CurrentTime time.Time `json:"currentTime"`
	TimeZone    struct {
		Id string `json:"id"`
	} `json:"timeZone"`
	Temperature struct {
		Degrees float64 `json:"degrees"`
		Unit    string  `json:"unit"`
	} `json:"temperature"`
	Precipitation struct {
		Probability struct {
			Percent int    `json:"percent"`
			Type    string `json:"type"`
		} `json:"probability"`
	} `json:"precipitation"`
}

func PopulateGoogleAPIData(DataCenter DataCenter) (WeatherData, error) {
	if GoogleApiKey == "" {
		return WeatherData{}, errors.New("GOOGLE_API_KEY not set")
	}

	geo, err := GetGeoCode(DataCenter.Address)
	if err != nil {
		return WeatherData{}, err
	}

	weather, err := GetWeather(geo)
	if err != nil {
		return WeatherData{}, err
	}

	return weather, nil
}

func GetGeoCode(address string) (GeoCode, error) {
	GoogleGeocodeBaseURL := "https://geocode.googleapis.com/v4/geocode/address/"

	requestURL := GoogleGeocodeBaseURL + url.PathEscape(address) + "?key=" + url.QueryEscape(GoogleApiKey)

	response, err := http.Get(requestURL)
	if err != nil {
		return GeoCode{}, err
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return GeoCode{}, err
	}

	var result GeocodeResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return GeoCode{}, err
	}

	if len(result.Results) > 0 {
		return result.Results[0], nil
	} else {
		return GeoCode{}, errors.New("no results found")
	}
}

func GetWeather(code GeoCode) (WeatherData, error) {
	GoogleWeatherBaseURL := "https://weather.googleapis.com/v1/currentConditions:lookup?"

	lat := strconv.FormatFloat(code.Location.Latitude, 'f', -1, 64)
	lng := strconv.FormatFloat(code.Location.Longitude, 'f', -1, 64)

	params := url.Values{}
	params.Set("key", GoogleApiKey)
	params.Set("location.latitude", lat)
	params.Set("location.longitude", lng)

	response, err := http.Get(GoogleWeatherBaseURL + params.Encode())
	if err != nil {
		fmt.Println(err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var result WeatherData
	err = json.Unmarshal(body, &result)
	if err != nil {
		return WeatherData{}, err
	}
	return result, nil
}
