package main

import (
	"fmt"
	"flag"
	"net/url"
	"strings"
	"github.com/gocolly/colly"
	. "github.com/logrusorgru/aurora"
)

func main() {
	// Define and parse command line flags
	domainPtr := flag.String("domain", "http://127.0.0.1", "Domain to crawl")
	depthPtr := flag.Int("depth", 1, "Maximum depth to crawl")
	includeJSPtr := flag.Bool("js", false, "Include links to utilised JavaScript files")
	flag.Parse()

	// These will store the discovered assets to avoid duplicates
	urls := make(map[string]struct{})
	subdomains := make(map[string]struct{})
	jsfiles := make(map[string]struct{})

	// The Colly collector
	c := colly.NewCollector(
		colly.MaxDepth(*depthPtr),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"),
	)

	// Find and visit the links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		var urlString string = e.Request.AbsoluteURL(e.Attr("href"))
		// If the url isn't already there, print and save it
		if _, ok := urls[urlString]; !ok {
			if urlString != ""{
				var urlObj, err = url.Parse(urlString) 
				if err != nil {
					fmt.Println(err)
				}
				if strings.Contains(urlObj.Host,*domainPtr){ 
					fmt.Println(BrightYellow("[url]"), urlString)
					urls[urlString] = struct{}{}
				}
				// If this is a new subdomain, print it
				if _, ok := subdomains[urlObj.Host]; !ok {
					if urlObj.Host != ""{
						if strings.Contains(urlObj.Host, *domainPtr){
							fmt.Println(BrightGreen("[subdomain]") , urlObj.Host)
							subdomains[urlObj.Host] = struct{}{}
						}
					}
				}
			}
			e.Request.Visit(e.Attr("href"))
		}
	})

	// Find and print all the JS files if "-js" is flagged
	if *includeJSPtr{
		c.OnHTML("script[src]", func(e *colly.HTMLElement) {
			jsfile := e.Request.AbsoluteURL(e.Attr("src"))
			if _, ok := jsfiles[jsfile]; !ok {
				if jsfile != ""{
					fmt.Println(BrightRed("[javascript]"), jsfile)
					jsfiles[jsfile] = struct{}{}
				}
			}
		})
	}

	c.Visit("http://" + *domainPtr)
	c.Visit("https://" + *domainPtr)
}
