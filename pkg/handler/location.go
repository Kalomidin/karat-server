package handler

import (
	"fmt"
	"time"
)

func loadLocation(name string) (*time.Location, error) {
	if name == "" {
		name = "UTC"
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("shop timezone: %w", err)
	}
	return location, nil
}
