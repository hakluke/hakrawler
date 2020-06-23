package config

import (
	"errors"
	"strings"
)

// Config represents the configuration for both the Collector and cli.
type Config struct {
	Url           string
	Depth         int
	Outdir        string
	Cookie        string
	AuthHeader    string
	Headers       string
	HeadersMap    map[string]string
	Scope         string
	Version       string
	Wayback       bool
	Plain         bool
	Nocolor			bool
	Runlinkfinder bool
	// output flags
	DisplayVersion bool
	IncludeJS      bool
	IncludeSubs    bool
	IncludeURLs    bool
	IncludeForms   bool
	IncludeRobots  bool
	IncludeSitemap bool
	IncludeWayback bool
	IncludeAll     bool
    Insecure       bool
}

// NewConfig returns a Config with default values.
func NewConfig() Config {
	var conf Config
	// default values
	conf.Version = "beta11"
	conf.DisplayVersion = false
	conf.Url = ""
	conf.Depth = 1
	conf.Outdir = ""
	conf.Cookie = ""
	conf.AuthHeader = ""
	conf.Headers = ""
	conf.HeadersMap = nil
	conf.Scope = "subs"
	conf.Wayback = false
	conf.Plain = false
	conf.Nocolor = false
	conf.Runlinkfinder = false
	// output flag default values
	conf.IncludeJS = false
	conf.IncludeSubs = false
	conf.IncludeURLs = false
	conf.IncludeForms = false
	conf.IncludeRobots = false
	conf.IncludeSitemap = false
	conf.IncludeWayback = false
	conf.IncludeAll = true
    conf.Insecure = false

	return conf
}

// VerifyFlags does validation and any manipulation that needs to happen from flags.
func VerifyFlags(conf *Config) error {
	if conf.Headers != "" {
		if !strings.Contains(conf.Headers, ":") {
			return errors.New("headers flag not formatted properly (no colon to separate header and value)")
		}

		headers := make(map[string]string)
		rawHeaders := strings.Split(conf.Headers, ";")
		for _, header := range rawHeaders {
			var parts []string
			if strings.Contains(header, ": ") {
				parts = strings.Split(header, ": ")
			} else if strings.Contains(header, ":") {
				parts = strings.Split(header, ":")
			} else {
				continue
			}
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
		conf.HeadersMap = headers
	}
	return nil
}
