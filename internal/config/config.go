package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config is how this process is run.
type Config struct {
	Addr            string
	DatabaseURL     string
	GoogleClientIDs []string
	ShopLocation    *time.Location
	DevLogin        bool
}

// FromEnv reads the process environment. A local .env file fills any value
// that is not already set in the environment.
func FromEnv() (Config, error) {
	loadDotEnv(".env")

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	zone := strings.TrimSpace(os.Getenv("SHOP_TIMEZONE"))
	if zone == "" {
		zone = "UTC"
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		return Config{}, fmt.Errorf("SHOP_TIMEZONE: %w", err)
	}

	var clientIDs []string
	for _, part := range strings.Split(os.Getenv("GOOGLE_CLIENT_IDS"), ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			clientIDs = append(clientIDs, part)
		}
	}
	dev := os.Getenv("DEV_LOGIN") == "1"
	if len(clientIDs) == 0 && !dev {
		return Config{}, fmt.Errorf("set GOOGLE_CLIENT_IDS, or DEV_LOGIN=1 for a local server")
	}

	addr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if addr == "" {
		addr = ":8080"
	}
	return Config{
		Addr:            addr,
		DatabaseURL:     databaseURL,
		GoogleClientIDs: clientIDs,
		ShopLocation:    location,
		DevLogin:        dev,
	}, nil
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
	}
}
