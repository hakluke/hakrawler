package hakrawler

type Config struct {
	Url           string
	Depth         int
	Outdir        string
	Cookie        string
	AuthHeader    string
	Scope         string
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

func NewConfig() Config {
	var conf Config
	// default values
	conf.Url = ""
	conf.Depth = 1
	conf.Outdir = ""
	conf.Cookie = ""
	conf.AuthHeader = ""
	conf.Scope = "subs"
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
