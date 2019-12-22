package main

import (
	"os"
	"fmt"
	"flag"
	"regexp"
	"net/url"
	"net/http"
	"strings"
	"io/ioutil"
	"github.com/gocolly/colly"
	. "github.com/logrusorgru/aurora"
	"github.com/yterajima/go-sitemap"
)

func banner(){
	fmt.Print(BrightRed(`
██╗  ██╗ █████╗ ██╗  ██╗██████╗  █████╗ ██╗    ██╗██╗     ███████╗██████╗ 
██║  ██║██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██║    ██║██║     ██╔════╝██╔══██╗
███████║███████║█████╔╝ ██████╔╝███████║██║ █╗ ██║██║     █████╗  ██████╔╝
██╔══██║██╔══██║██╔═██╗ ██╔══██╗██╔══██║██║███╗██║██║     ██╔══╝  ██╔══██╗
██║  ██║██║  ██║██║  ██╗██║  ██║██║  ██║╚███╔███╔╝███████╗███████╗██║  ██║
╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚══╝╚══╝ ╚══════╝╚══════╝╚═╝  ╚═╝
`))
	fmt.Println(BgBlue(BrightYellow("                        Crafted with <3 by hakluke                        ")))
	fmt.Println(" ")
}

func colorPrint(tag Value, msg string){
	fmt.Println(tag, msg)
}

func printIfInScope(scope string, tag Value, domain string, msg string){
	var urlObj, err = url.Parse(msg)
	if err != nil {
		fmt.Println(err)
	}
	switch scope {
	case "strict":
		if urlObj.Host == domain {
			colorPrint(tag, msg)
		}
	case "subs":
		if strings.Contains(urlObj.Host, domain) {
			colorPrint(tag, msg)
		}
	default:
		colorPrint(tag, msg)
	}
}

func parseSitemap(domain string, depth int, c colly.Collector, printResult bool){
	schemas := [2]string{"http://", "https://"}
	var sitemapurls []string
	for _, schema := range schemas {
		sitemapURL := schema+domain+"/sitemap.xml"
		smap, err := sitemap.Get(sitemapURL, nil)	
		if err!=nil {
			fmt.Println(err)
			break	
		}
		for _, URL := range smap.URL {
			if(printResult){
				fmt.Println(BrightBlue("[sitemap]"), URL.Loc)
			}
			//add it to a slice for parsing later
			sitemapurls = append(sitemapurls, URL.Loc)
		}
	}

	// if depth is greater than 1, add all of the sitemap urls as seeds
	if depth > 1 {
		for _, sitemapurl := range sitemapurls {
			c.Visit(sitemapurl)
		}
	}
}

func parseRobots(domain string, depth int, c colly.Collector, printResult bool){
	schemas := [2]string{"http://", "https://"}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
	}}
	var robotsurls []string
	for _, schema := range schemas {
		robotsURL := schema+domain+"/robots.txt"
		nextURL := robotsURL

		for {
			resp, err := client.Get(nextURL)	
			if err!=nil {
				fmt.Println(err)
				break	
			}
			if schema == "http://" && strings.Contains(resp.Header.Get("Location"), "https://"){
				break
			}
			if resp.StatusCode == 200{
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil{
					print(err)
					return
				}
				lines := strings.Split(string(body), "\n")

				var re = regexp.MustCompile(".*llow: ")
				for _, line := range lines {
					if strings.Contains(line, "llow: "){
						urlstring := re.ReplaceAllString(line, "")
						if(printResult){
							fmt.Println(BrightMagenta("[robots]"), schema + domain + urlstring)
						}
						//add it to a slice for parsing later
						robotsurls = append(robotsurls, schema + domain + urlstring)
					}
				}
				break
			} else {
				nextURL = resp.Header.Get("Location")
			}
		}
	}
	// if depth is greater than 1, add all of the robots urls as seeds
	if depth > 1 {
		for _, robotsurl := range robotsurls{
			c.Visit(robotsurl)
		}
	}
}

func main() {

	// print the banner
	banner()

	// define and parse command line flags
	domainPtr := flag.String("domain", "", "Domain to crawl")
	depthPtr := flag.Int("depth", 1, "Maximum depth to crawl")

	// which data to include in output?
	includeJSPtr := flag.Bool("js", false, "Include links to utilised JavaScript files")
	includeSubsPtr := flag.Bool("subs", false, "Include subdomains")
	includeURLsPtr := flag.Bool("urls", false, "Include URLs")
	includeFormsPtr := flag.Bool("forms", false, "Include form actions")
	includeRobotsPtr := flag.Bool("robots", false, "Include robots.txt entries")
	includeSitemapPtr := flag.Bool("sitemap", false, "Include sitemap.xml entries")
	includeAllPtr := flag.Bool("all", true, "Include everything")
	scopePtr := flag.String("scope", "loose", "Scope to include:\nstrict = specified domain only\nsubs = specified domain and subdomains\nloose = everything")

	flag.Parse()

	// make sure the domain has been set
	if *domainPtr == "" {
		fmt.Println(BrightRed("[error]"), "You must set a domain, e.g. -domain=example.com")
		fmt.Println(BrightBlue("[info]"), "See hakrawler -h for commandline options")
		os.Exit(1)
	}

	// set up the bools
	if *includeJSPtr || *includeSubsPtr || *includeURLsPtr || *includeFormsPtr || *includeRobotsPtr || *includeSitemapPtr {
		*includeAllPtr = false
	}

	// these will store the discovered assets to avoid duplicates
	urls := make(map[string]struct{})
	subdomains := make(map[string]struct{})
	jsfiles := make(map[string]struct{})
	forms := make(map[string]struct{})
	//robots := make(map[string]struct{}) // probably don't need this...

	// behold, the colly collector
	c := colly.NewCollector(
		colly.MaxDepth(*depthPtr),
		// this is not fooling anyone
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"),
	)

	// find and visit the links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		var urlString string = e.Request.AbsoluteURL(e.Attr("href"))
		// if the url isn't already there, print and save it, if it's a new subdomain, print that too
		if _, ok := urls[urlString]; !ok {
			if urlString != ""{
				var urlObj, err = url.Parse(urlString) 
				if err != nil {
					fmt.Println(err)
				}
				if *includeURLsPtr || *includeAllPtr {
					printIfInScope(*scopePtr,BrightYellow("[url]"),*domainPtr,urlString)
					urls[urlString] = struct{}{}
				}
				// if this is a new subdomain, print it
				if *includeSubsPtr || *includeAllPtr {
					if _, ok := subdomains[urlObj.Host]; !ok {
						if urlObj.Host != ""{
							if strings.Contains(urlObj.Host, *domainPtr){
								printIfInScope(*scopePtr,BrightGreen("[subdomain]"),*domainPtr,urlObj.Host)
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
	if *includeJSPtr || *includeAllPtr {
		c.OnHTML("script[src]", func(e *colly.HTMLElement) {
			jsfile := e.Request.AbsoluteURL(e.Attr("src"))
			if _, ok := jsfiles[jsfile]; !ok {
				if jsfile != ""{
					printIfInScope(*scopePtr,BrightRed("[javascript]"),*domainPtr,jsfile)
					jsfiles[jsfile] = struct{}{}
				}
			}
		})
	}


	// find and print all the form action URLs 
	if *includeFormsPtr || *includeAllPtr {
		c.OnHTML("form[action]", func(e *colly.HTMLElement) {
			form := e.Request.AbsoluteURL(e.Attr("action"))
			if _, ok := forms[form]; !ok {
				if form != ""{
					printIfInScope(*scopePtr,BrightCyan("[form]"),*domainPtr,form)
					forms[form] = struct{}{}
				}
			}
		})
	}

	// figure out if the results from robots.txt should be printed
	var printRobots bool
    if *includeRobotsPtr || *includeAllPtr {
		printRobots = true	
	} else {
		printRobots = false	
	}	

	// figure out of the results from sitemap.xml should be printed
	var printSitemap bool
	if *includeSitemapPtr || *includeAllPtr {
		printSitemap = true
	} else {
		printSitemap = false
	}

	// do all the things
	parseRobots(*domainPtr, *depthPtr, *c, printRobots)
	parseSitemap(*domainPtr, *depthPtr, *c, printSitemap)
	c.Visit("http://" + *domainPtr)
	c.Visit("https://" + *domainPtr)
}
