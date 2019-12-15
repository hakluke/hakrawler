package main

import (
	"fmt"
	"flag"
	"github.com/gocolly/colly"
)

func main() {
	// Define and parse command line flags
	domainPtr := flag.String("domain", "http://127.0.0.1", "Domain to crawl")
	depthPtr := flag.Int("depth", 1, "Maximum depth to crawl")
	includeJSPtr := flag.Bool("js", false, "Include links to utilised JavaScript files")
	flag.Parse()

	var domain string = *domainPtr 

	urls := make(map[string]struct{})

	c := colly.NewCollector(
		colly.AllowedDomains(domain, "www." + domain),
		colly.MaxDepth(*depthPtr),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		var url string = e.Request.AbsoluteURL(e.Attr("href"))
		// If the url isn't already there, print and save it
		if _, ok := urls[url]; !ok {
			if url != ""{
				fmt.Println(url)
				urls[url] = struct{}{}
			}
		}
		//fmt.Println(e.Request.AbsoluteURL(e.Attr("href")))
		e.Request.Visit(e.Attr("href"))
	})

	// Find and print all the JS files if "-js" is flagged
	if *includeJSPtr{
		c.OnHTML("script[src]", func(e *colly.HTMLElement) {
			fmt.Println(e.Request.AbsoluteURL(e.Attr("src")))
		})
	}

	// print each URL that is being visited
	//c.OnRequest(func(r *colly.Request) {
	//	fmt.Println("### VISITING ", r.URL)
	//})

	c.Visit("http://" + *domainPtr)
	c.Visit("https://" + *domainPtr)

}

