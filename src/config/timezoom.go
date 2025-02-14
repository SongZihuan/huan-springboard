package config

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var location *time.Location
var locationOnce *sync.Once = new(sync.Once)

func TimeZone() *time.Location {
	locationOnce.Do(func() {
		if !IsReady() {
			panic("config not ready")
		}

		if strings.ToLower(config.GetConfig().Timezone) == "utc" {
			_location := time.UTC
			if _location == nil {
				_location = time.Local
			}

			if _location != nil {
				location = _location
			}
		} else if strings.ToLower(config.GetConfig().Timezone) == "local" || config.GetConfig().Timezone == "" {
			_location := time.Local
			if _location == nil {
				_location = time.UTC
			}

			if _location != nil {
				location = _location
			}
		} else {
			_location, err := time.LoadLocation(config.GetConfig().Timezone)
			if err != nil || _location == nil {
				_location = time.UTC
			}

			if _location != nil {
				location = _location
			}
		}

		if location == nil {
			if config.GetConfig().Timezone == "UTC" || config.GetConfig().Timezone == "Local" || config.GetConfig().Timezone == "" {
				panic(fmt.Errorf("can not get location UTC or Local"))
			}
			panic(fmt.Errorf("can not get location UTC, Local or %s", config.GetConfig().Timezone))
		}
	})

	return location
}
