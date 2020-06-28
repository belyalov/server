package sun

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
)

var flagLat = flag.Float64("sun.lat", 39.575295, "Latitude")
var flagLong = flag.Float64("sun.long", -104.902129, "Longitude")
var flagInterval = flag.String("sun.interval", "4h", "Sun data refresh interval")

var mutex sync.Mutex
var currentSunData sunData
var apiURL string

const (
	apiURLTemplate = "https://api.sunrise-sunset.org/json?lat=%f&lng=%f&formatted=0"
)

type sunData struct {
	// JSONResults maps to "results" field in JSON repply
	JSONResults JSONResults `json:"results"`
	Status      string

	Sunset    time.Time
	Sunrise   time.Time
	SolarNoon time.Time
	DayLength time.Duration
}

// JSONResults represents sunrise-sunset API
type JSONResults struct {
	SunriseString   string `json:"sunrise"`
	SunsetString    string `json:"sunset"`
	SolarNoonString string `json:"solar_noon"`
	DayLengthInt    int    `json:"day_length"`
}

// GetSunset returns current's day time of sunset
// Data maybe stale
func GetSunset() time.Time {
	mutex.Lock()
	defer mutex.Unlock()

	return currentSunData.Sunset
}

// GetSunrise returns current's day time of sunrise
// Data maybe stale
func GetSunrise() time.Time {
	mutex.Lock()
	defer mutex.Unlock()

	return currentSunData.Sunrise
}

// Start starts periodical sun data refresher
// It also updates data right away
func Start(ctx context.Context) error {
	// Prepare / parse parameters
	apiURL = fmt.Sprintf(apiURLTemplate, *flagLat, *flagLong)
	interval, err := time.ParseDuration(*flagInterval)
	if err != nil {
		return err
	}

	// Run first data refresh immediately
	if err := refreshData(); err != nil {
		return err
	}

	// .. and on periodical bases
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				err := refreshData()
				if err != nil {
					glog.Errorf("Unable to refresh sun data: %v", err)
				}
			case <-ctx.Done():
				glog.Infof("Sun updater terminated")
			}
		}
	}()

	return nil
}

func refreshData() error {
	// Perform HTTP request to get updated sun data
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse JSON
	var result sunData
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&result)
	if result.Status != "OK" {
		return fmt.Errorf("%s wrong return status code: %v", apiURL, result.Status)
	}

	// Convert dates to golang format
	result.Sunset = convertDate(result.JSONResults.SunsetString)
	result.Sunrise = convertDate(result.JSONResults.SunriseString)
	result.SolarNoon = convertDate(result.JSONResults.SolarNoonString)
	result.DayLength, _ = time.ParseDuration(fmt.Sprintf("%ds", result.JSONResults.DayLengthInt))
	glog.Infof("Got new sun data: sunrise %s, sunset %s, solar noon %s, day lenght %s",
		result.Sunset,
		result.Sunrise,
		result.SolarNoon,
		result.DayLength,
	)

	// Update cache
	mutex.Lock()
	defer mutex.Unlock()
	currentSunData = result

	return nil
}

func convertDate(input string) time.Time {
	parsed, _ := time.Parse(time.RFC3339, input)
	return parsed.In(time.Local)
}
