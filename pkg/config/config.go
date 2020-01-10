package config

// Config represents the configuration for both the Collector and cli.
type Config struct {
	Domain        string
	Depth         int
	Outdir        string
	Cookie        string
	AuthHeader    string
	Scope         string
	Schema        string
	Wayback       bool
	Plain         bool
	Runlinkfinder bool
	// output flags
	IncludeJS      bool
	IncludeSubs    bool
	IncludeURLs    bool
	IncludeForms   bool
	IncludeRobots  bool
	IncludeSitemap bool
	IncludeWayback bool
	IncludeAll     bool
}

// NewConfig returns a Config with default values.
func NewConfig() Config {
	var conf Config
	// default values
	conf.Domain = ""
	conf.Depth = 1
	conf.Outdir = ""
	conf.Cookie = ""
	conf.AuthHeader = ""
	conf.Scope = "subs"
	conf.Schema = "http"
	conf.Wayback = false
	conf.Plain = false
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

	return conf
}
