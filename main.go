package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/hakluke/hakrawler/pkg/hakrawler"

	"github.com/gocolly/colly"
	"github.com/google/uuid"
	. "github.com/logrusorgru/aurora"
	sitemap "github.com/oxffaa/gopher-parse-sitemap"
)

var (
	// regular expression from LinkFinder.
	LinkFinderRegex, _ = regexp.Compile(`(?:"|')(((?:[a-zA-Z]{1,10}://|//)[^"'/]{1,}\.[a-zA-Z]{2,}[^"']{0,})|((?:/|\.\./|\./)[^"'><,;| *()(%%$^/\\\[\]][^"'><,;|()]{1,})|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{1,}\.(?:[a-zA-Z]{1,4}|action)(?:[\?|/][^"|']{0,}|))
|([a-zA-Z0-9_\-]{1,}\.(?:php|asp|aspx|jsp|json|action|html|js|txt|xml)(?:\?[^"|']{0,}|)))(?:"|')`)
)

func banner(au Aurora) {
	fmt.Print(au.BrightRed(`
██╗  ██╗ █████╗ ██╗  ██╗██████╗  █████╗ ██╗    ██╗██╗     ███████╗██████╗
██║  ██║██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██║    ██║██║     ██╔════╝██╔══██╗
███████║███████║█████╔╝ ██████╔╝███████║██║ █╗ ██║██║     █████╗  ██████╔╝
██╔══██║██╔══██║██╔═██╗ ██╔══██╗██╔══██║██║███╗██║██║     ██╔══╝  ██╔══██╗
██║  ██║██║  ██║██║  ██╗██║  ██║██║  ██║╚███╔███╔╝███████╗███████╗██║  ██║
╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚══╝╚══╝ ╚══════╝╚══════╝╚═╝  ╚═╝
`))
	fmt.Println(BgBlue(au.BrightYellow("                        Crafted with <3 by hakluke                        ")))
}

// if -plain is set, just print the message, otherwise print a coloured tag and then the message
func colorPrint(tag Value, msg string, plain bool) {
	if plain {
		fmt.Println(msg)
	} else {
		fmt.Println(tag, msg)
	}
}

func printToRandomFile(msg string, dir string) {
	uuid, _ := uuid.NewRandom()
	if dir[len(dir)-1:] != "/" {
		dir = dir + "/"
	}

	err := ioutil.WriteFile(dir+"hakrawler_"+uuid.String()+".req", []byte(msg), 0644)
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
}

// determines whether the domains/urls should be printed based on the provided scope (returns true/false)
func printIfInScope(conf hakrawler.Config, tag Value, schema string, msg string) bool {
	base, err := url.Parse(schema + conf.Domain)
	if err != nil {
		// Error parsing base domain
		return false
	}

	msgSchema := ""
	if !strings.Contains(msg, "http://") && !strings.Contains(msg, "https://") && !strings.HasPrefix(msg, "/") {
		msgSchema = "https://"
	}

	urlObj, err := url.Parse(msgSchema + msg)
	if err != nil {
		// the url can't be parsed, move on with reckless abandon
		return false
	}
	urlObj = base.ResolveReference(urlObj)

	shouldPrint := false

	switch conf.Scope {
	case "strict":
		shouldPrint = urlObj.Host == conf.Domain
	case "fuzzy":
		shouldPrint = strings.Contains(urlObj.Host, conf.Domain)
	case "subs":
		shouldPrint = strings.HasSuffix(urlObj.Host, conf.Domain)
	default:
		shouldPrint = true
	}

	if !shouldPrint {
		return false
	}

	strVal := urlObj.String()
	// Remove the schema if it was added before
	if msgSchema != "" {
		strVal = strings.Replace(strVal, msgSchema, "", 1)
	}
	colorPrint(tag, strVal, conf.Plain)
	if conf.Outdir != "" {
		printToRandomFile(rawHTTPGET(msg), conf.Outdir)
	}
	return true
}

func rawHTTPGET(url string) string {
	// some sanity checking
	if !strings.Contains(url, "http") {
		return ""
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	raw, err := httputil.DumpRequest(req, true)
	if err != nil {
		return ""
	}
	return string(raw)
}

func parseSitemap(conf hakrawler.Config, c colly.Collector, printResult bool, mainwg *sync.WaitGroup, schema string, au Aurora) {
	defer mainwg.Done()
	sitemapURL := schema + conf.Domain + "/sitemap.xml"
	sitemap.ParseFromSite(sitemapURL, func(e sitemap.Entry) error {
		if printResult {
			_ = printIfInScope(conf, au.BrightBlue("[sitemap]"), schema, e.GetLocation())
		}
		// if depth is greater than 1, add sitemap url as seed
		if conf.Depth > 1 {
			c.Visit(e.GetLocation())
		}
		return nil
	})
}

func parseRobots(conf hakrawler.Config, c colly.Collector, printResult bool, mainwg *sync.WaitGroup, schema string, au Aurora) {
	defer mainwg.Done()
	var robotsurls []string
	robotsURL := schema + conf.Domain + "/robots.txt"

	resp, err := http.Get(robotsURL)
	if err != nil {
		return
	}
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		lines := strings.Split(string(body), "\n")

		var re = regexp.MustCompile(".*llow: ")
		for _, line := range lines {
			if strings.Contains(line, "llow: ") {
				urlstring := re.ReplaceAllString(line, "")
				if printResult {
					_ = printIfInScope(conf, au.BrightMagenta("[robots]"), schema, schema+conf.Domain+urlstring)
				}
				//add it to a slice for parsing later
				robotsurls = append(robotsurls, schema+conf.Domain+urlstring)
			}
		}
	}
	// if depth is greater than 1, add all of the robots urls as seeds
	if conf.Depth > 1 {
		for _, robotsurl := range robotsurls {
			c.Visit(robotsurl)
		}
	}
}
func linkfinder(jsfile string, tag Value, plain bool) {
	// skip tls verification
	client := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(jsfile)
	if err != nil {
		return
	}
	//  if js file exists
	if resp.StatusCode == 200 {
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		resp.Body.Close()
		found := LinkFinderRegex.FindAllString(string(res), -1)
		for _, link := range found {
			colorPrint(tag, link, plain)
		}

	}
}
func crawl(conf hakrawler.Config, au Aurora, domainwg *sync.WaitGroup) {

	// make sure the domain has been set
	if conf.Domain == "" {
		fmt.Println(au.BrightRed("[error]"), "You must set a domain, e.g. -domain=example.com")
		fmt.Println(au.BrightBlue("[info]"), "See hakrawler -h for commandline options")
		os.Exit(1)
	}

	// set up the schema
	schema := "http://"
	if conf.Schema == "https" {
		schema = "https://"
	}

	// these will store the discovered assets to avoid duplicates
	urls := make(map[string]struct{})
	subdomains := make(map[string]struct{})
	jsfiles := make(map[string]struct{})
	forms := make(map[string]struct{})

	// behold, the colly collector
	c := colly.NewCollector(
		colly.MaxDepth(conf.Depth),
		// this is not fooling anyone XD
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"),
		//colly.Async(true),
	)

	// set custom headers if specified
	if conf.Cookie != "" {
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Cookie", conf.Cookie)
		})
	}

	if conf.AuthHeader != "" {
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Authorization", conf.AuthHeader)
		})
	}

	// find and visit the links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		urlString := e.Request.AbsoluteURL(e.Attr("href"))
		// if the url isn't already there, print and save it, if it's a new subdomain, print that too
		if _, ok := urls[urlString]; !ok {
			if urlString != "" {
				urlObj, err := url.Parse(urlString)
				// ditch unparseable URLs
				if err != nil {
					fmt.Println(err)
				} else {
					if conf.IncludeURLs || conf.IncludeAll {
						_ = printIfInScope(conf, au.BrightYellow("[url]"), schema, urlString)
						urls[urlString] = struct{}{}
					}
					// if this is a new subdomain, print it
					if conf.IncludeSubs || conf.IncludeAll {
						if _, ok := subdomains[urlObj.Host]; !ok {
							if urlObj.Host != "" {
								_ = printIfInScope(conf, au.BrightGreen("[subdomain]"), schema, urlObj.Host)
								subdomains[urlObj.Host] = struct{}{}
							}
						}
					}
				}
			}
			e.Request.Visit(e.Attr("href"))
		}
	})

	// find and print all the JavaScript files
	if conf.IncludeJS || conf.IncludeAll {
		c.OnHTML("script[src]", func(e *colly.HTMLElement) {
			jsfile := e.Request.AbsoluteURL(e.Attr("src"))
			if _, ok := jsfiles[jsfile]; !ok {
				if jsfile != "" {
					inScope := printIfInScope(conf, au.BrightRed("[javascript]"), schema, jsfile)
					if inScope {
						if conf.Runlinkfinder {
							linkfinder(jsfile, au.BrightRed("[linkfinder]"), conf.Plain)
						}
					}
					jsfiles[jsfile] = struct{}{}
				}
			}
		})
	}

	// find and print all the form action URLs
	if conf.IncludeForms || conf.IncludeAll {
		c.OnHTML("form[action]", func(e *colly.HTMLElement) {
			form := e.Request.AbsoluteURL(e.Attr("action"))
			if _, ok := forms[form]; !ok {
				if form != "" {
					_ = printIfInScope(conf, au.BrightCyan("[form]"), schema, form)
					forms[form] = struct{}{}
				}
			}
		})
	}

	// figure out if the results from robots.txt should be printed
	printRobots := conf.IncludeRobots || conf.IncludeAll

	// figure out of the results from sitemap.xml should be printed
	printSitemap := conf.IncludeSitemap || conf.IncludeAll

	// do all the things
	// setup a waitgroup to run all methods at the same time
	var mainwg sync.WaitGroup

	// robots.txt
	mainwg.Add(1)
	go parseRobots(conf, *c, printRobots, &mainwg, schema, au)

	// sitemap.xml
	mainwg.Add(1)
	go parseSitemap(conf, *c, printSitemap, &mainwg, schema, au)

	// waybackurls
	if conf.Wayback {
		go func() {
			mainwg.Add(1)
			defer mainwg.Done()
			// get results from waybackurls
			waybackurls := WaybackURLs(conf.Domain)

			// print wayback results, if depth >1, also add them to the crawl queue
			for _, waybackurl := range waybackurls {
				if conf.IncludeWayback || conf.IncludeAll {
					_ = printIfInScope(conf, au.Yellow("[wayback]"), schema, waybackurl)
				}
				// if this is a new subdomain, print it
				urlObj, err := url.Parse(waybackurl)
				if err != nil {
					continue
				}

				if conf.IncludeSubs || conf.IncludeAll {
					if _, ok := subdomains[urlObj.Host]; !ok {
						if urlObj.Host != "" {
							if strings.Contains(urlObj.Host, conf.Domain) {
								_ = printIfInScope(conf, au.BrightGreen("[subdomain]"), schema, urlObj.Host)
								subdomains[urlObj.Host] = struct{}{}
							}
						}
					}
				}
				if conf.Depth > 1 {
					c.Visit(waybackurl)
				}
			}
		}()
	}

	// colly
	mainwg.Add(1)
	go func() {
		defer mainwg.Done()
		c.Visit(schema + conf.Domain)
	}()

	mainwg.Wait()
	domainwg.Done()
}

func main() {
	conf := hakrawler.NewConfig()
	// define and parse command line flags
	flag.StringVar(&conf.Domain, "domain", "", "The domain that you wish to crawl (for example, google.com)")
	flag.IntVar(&conf.Depth, "depth", 1, "Maximum depth to crawl, the default is 1. Anything above 1 will include URLs from robots, sitemap, waybackurls and the initial crawler as a seed. Higher numbers take longer but yield more results.")
	flag.StringVar(&conf.Outdir, "outdir", "", "Directory to save discovered raw HTTP requests")
	flag.StringVar(&conf.Cookie, "cookie", "", "The value of this will be included as a Cookie header")
	flag.StringVar(&conf.AuthHeader, "auth", "", "The value of this will be included as a Authorization header")
	flag.StringVar(&conf.Scope, "scope", "subs", "Scope to include:\nstrict = specified domain only\nsubs = specified domain and subdomains\nfuzzy = anything containing the supplied domain\nyolo = everything")
	flag.StringVar(&conf.Schema, "schema", "http", "Schema, http or https")
	flag.BoolVar(&conf.Wayback, "usewayback", false, "Query wayback machine for URLs and add them as seeds for the crawler")
	flag.BoolVar(&conf.Plain, "plain", false, "Don't use colours or print the banners to allow for easier parsing")
	flag.BoolVar(&conf.Runlinkfinder, "linkfinder", false, "Run linkfinder on javascript files.")

	// which data to include in output?
	flag.BoolVar(&conf.IncludeJS, "js", false, "Include links to utilised JavaScript files")
	flag.BoolVar(&conf.IncludeSubs, "subs", false, "Include subdomains in output")
	flag.BoolVar(&conf.IncludeURLs, "urls", false, "Include URLs in output")
	flag.BoolVar(&conf.IncludeForms, "forms", false, "Include form actions in output")
	flag.BoolVar(&conf.IncludeRobots, "robots", false, "Include robots.txt entries in output")
	flag.BoolVar(&conf.IncludeSitemap, "sitemap", false, "Include sitemap.xml entries in output")
	flag.BoolVar(&conf.IncludeWayback, "wayback", false, "Include wayback machine entries in output")
	flag.BoolVar(&conf.IncludeAll, "all", true, "Include everything in output - this is the default, so this option is superfluous")
	flag.Parse()

	// set up the bools
	if conf.IncludeJS || conf.IncludeSubs || conf.IncludeURLs || conf.IncludeForms || conf.IncludeRobots || conf.IncludeSitemap {
		conf.IncludeAll = false
	}

	au := NewAurora(!conf.Plain)

	// print the banner
	if !conf.Plain {
		banner(au)
	}

	// decide whether to use -domain or stdin
	var domainwg sync.WaitGroup
	if conf.Domain != "" {
		// in this case, the waitgroup is not necessary as there is only 1 domain
		// I added it anyway because the function is expecting a wg pointer
		// There's a better way to do this
		domainwg.Add(1)
		go crawl(conf, au, &domainwg)
	} else {
		// get domains from stdin
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			domainwg.Add(1)
			conf.Domain = strings.ToLower(sc.Text())
			go crawl(conf, au, &domainwg)
		}
	}
	domainwg.Wait()
}
