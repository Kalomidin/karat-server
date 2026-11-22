package google

import "strings"

type Config struct {
	Audiences string `yaml:"audiences" env:"GOOGLE_CLIENT_IDS"`
	DevLogin  bool   `yaml:"devLogin" env:"DEV_LOGIN" env-default:"false"`
}

func (cfg Config) AudienceList() []string {
	var audiences []string
	for _, part := range strings.Split(cfg.Audiences, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			audiences = append(audiences, part)
		}
	}
	return audiences
}

type Client struct {
	cfg Config
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg}
}
