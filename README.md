# weather

Fetches current weather for a set of data-center locations by chaining two
Google APIs: first turning a street address into coordinates, then looking up
the live conditions at those coordinates. Lookups run **concurrently**, one
goroutine per location.

## How it works

For each location the program runs a two-step chain:

```
Address ──GetGeoCode──▶ GeoCode (lat/lng) ──GetWeather──▶ WeatherData
```

1. **`GetGeoCode`** calls the Google Geocoding API v4
   (`geocode.googleapis.com/v4/geocode/address/`) and deserialises the
   `results[0]` object into a `GeoCode` (place ID + latitude/longitude).
2. **`GetWeather`** calls the Google Weather API
   (`weather.googleapis.com/v1/currentConditions:lookup`) with those
   coordinates and deserialises the response into a `WeatherData`
   (temperature, precipitation probability, timezone).
3. **`PopulateGoogleAPIData`** ties the two together for a single
   `DataCenter`, returning the populated `WeatherData`.

The two steps are **dependent** — `GetWeather` needs the coordinates that
`GetGeoCode` produced — so within one location they run in sequence. Different
locations are **independent**, so they run in parallel.

## Types

| Type                 | Purpose                                                         |
|----------------------|-----------------------------------------------------------------|
| `DataCenter`         | A location: its address plus the geocode and weather we fill in |
| `GeoCode` / `LatLng` | Coordinates returned by the geocoding API                       |
| `WeatherData`        | Current conditions returned by the weather API                  |
| `GeocodeResponse`    | Wrapper matching the `{ "results": [...] }` JSON envelope       |

## Requirements

- Go 1.22 or newer (relies on per-iteration loop-variable semantics)
- A Google API key with the **Geocoding API** and **Weather API** enabled

## Setup

Set your API key in the environment — the program reads it at startup:

```bash
export GOOGLE_API_KEY="your-key-here"
```

## Running

The concurrent fan-out currently lives in the test, which doubles as a
profiling / run harness. Run it with verbose output so you see the per-location
results and logs:

```bash
go test -v ./...
```

Always run with the race detector while working on the concurrency:

```bash
go test -race -v ./...
```

If `GOOGLE_API_KEY` is unset, every lookup fails fast with
`GOOGLE_API_KEY not set`.

## Concurrency model

The fan-out uses three pieces working together:

- **A goroutine per location** runs the geocode→weather chain.
- **A `sync.WaitGroup`** counts the in-flight goroutines so we know when the
  last one has finished.
- **A buffered channel** (`chan WeatherData`) collects each result, and a
  dedicated closer goroutine closes it once the WaitGroup hits zero — which is
  what ends the receiving `range` loop.

See the walkthrough below for line-by-line detail.