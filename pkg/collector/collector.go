package collector

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/hakluke/hakrawler/pkg/config"
	sitemap "github.com/oxffaa/gopher-parse-sitemap"

	"github.com/logrusorgru/aurora"
)

// Collector exposes functions to scrape web pages and write results to a writer.
type Collector struct {
	conf  *config.Config
	colly *colly.Collector
	au    aurora.Aurora
	w     io.Writer
}

// NewCollector returns an initialized Collector.
func NewCollector(config *config.Config, au aurora.Aurora, w io.Writer) *Collector {
	c := colly.NewCollector(
		colly.MaxDepth(config.Depth),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"),
	)
	// set custom headers if specified
	if config.Cookie != "" {
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Cookie", config.Cookie)
		})
	}

	if config.AuthHeader != "" {
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Authorization", config.AuthHeader)
		})
	}
	return &Collector{
		conf:  config,
		colly: c,
		au:    au,
		w:     w,
	}
}

type syncList struct {
	L    sync.Mutex
	Reqs []*http.Request
}

func (s *syncList) AddReq(r *http.Request) {
	s.L.Lock()
	defer s.L.Unlock()
	s.Reqs = append(s.Reqs, r)
}

// Crawl crawls a domain for urls, subdomains, jsfiles, and forms, printing output as it goes. Returns a list of http requests made
func (c *Collector) Crawl(url string) ([]*http.Request, error) {
	// make sure the url has been set
	if url == "" {
		return []*http.Request{}, errors.New("url was empty")
	}

	// these will store the discovered assets to avoid duplicates
	var urls sync.Map
	var subdomains sync.Map
	var jsfiles sync.Map
	var forms sync.Map

	//urls := make(map[string]struct{})
	//subdomains := make(map[string]struct{})
	//jsfiles := make(map[string]struct{})
	//forms := make(map[string]struct{})
	reqsMade := &syncList{}

	// find and visit the links
	c.colly.OnHTML("a[href]", c.visitHTMLFunc(&urls, &subdomains, url, reqsMade))

	if c.conf.IncludeJS || c.conf.IncludeAll {
		// find and print all the JavaScript files
		c.colly.OnHTML("script[src]", c.findJSFunc(&jsfiles, url, reqsMade))
	}

	if c.conf.IncludeForms || c.conf.IncludeAll {
		// find and print all the form action URLs
		c.colly.OnHTML("form[action]", c.findFormsFunc(&forms, url, reqsMade))
	}

	// setup a waitgroup to run all methods at the same time
	var wg sync.WaitGroup

	// robots.txt
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.parseRobots(url, reqsMade)
	}()

	// sitemap
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.parseSitemap(url, reqsMade)
	}()

	// waybackurls
	if c.conf.Wayback {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.visitWaybackURLs(url, &subdomains, reqsMade)
		}()
	}

	// colly
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.colly.Visit(url)
	}()

	wg.Wait()
	return reqsMade.Reqs, nil
}

func (c *Collector) visitHTMLFunc(urls *sync.Map, subdomains *sync.Map, u string, reqsMade *syncList) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		var urlString string = e.Request.AbsoluteURL(e.Attr("href"))
		// if the url isn't already there, print and save it, if it's a new subdomain, print that too
		if _, exists := urls.Load(urlString); exists {
			return
		}
		if urlString == "" {
			return
		}
		var urlObj, err = url.Parse(urlString)

		if err != nil {
			return
		}

		e.Request.Visit(e.Attr("href"))
		if c.conf.IncludeURLs || c.conf.IncludeAll {
			c.recordIfInScope(c.au.BrightYellow("[url]"), u, urlString, reqsMade)
			urls.Store(urlString, struct{}{})
		}

		if !c.conf.IncludeSubs && !c.conf.IncludeAll {
			return
		}

		// if this is a new subdomain, print it
		if _, exists := subdomains.Load(urlObj.Host); exists {
			return
		}

		if urlObj.Host != "" {
			c.recordIfInScope(c.au.BrightGreen("[subdomain]"), u, urlObj.Host, reqsMade)
			subdomains.Store(urlObj.Host, struct{}{})
		}
	}
}

func (c *Collector) findJSFunc(jsfiles *sync.Map, u string, reqsMade *syncList) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		jsfile := e.Request.AbsoluteURL(e.Attr("src"))
		if _, exists := jsfiles.Load(jsfile); exists {
			return
		}
		if jsfile == "" {
			return
		}
		inScope := c.recordIfInScope(c.au.BrightRed("[javascript]"), u, jsfile, reqsMade)
		if inScope && c.conf.Runlinkfinder {
			c.linkfinder(jsfile, c.au.BrightRed("[linkfinder]"), c.conf.Plain)
		}
		jsfiles.Store(jsfile, struct{}{})
	}
}

func (c *Collector) findFormsFunc(forms *sync.Map, u string, reqsMade *syncList) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		form := e.Request.AbsoluteURL(e.Attr("action"))
		if _, exists := forms.Load(form); exists {
			return
		}
		if form == "" {
			return
		}
		c.recordIfInScope(c.au.BrightCyan("[form]"), u, form, reqsMade)
		forms.Store(form, struct{}{})
	}
}

func parseHostFromURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	return parsed.Host, nil
}

// recordIfInScope determines whether the domains/urls should be printed based on the provided scope (returns true/false).
func (c *Collector) recordIfInScope(tag aurora.Value, u string, msg string, reqsMade *syncList) bool {
	basehost, err := parseHostFromURL(u)
	if err != nil {
		// Error parsing base domain
		return false
	}

	msgHost := msg
	if strings.Contains(msg, "://") {
		msgHost, err = parseHostFromURL(msg)
		if err != nil {
			return false
		}
	}

	var shouldPrint bool

	switch c.conf.Scope {
	case "strict":
		shouldPrint = msgHost == basehost
	case "fuzzy":
		shouldPrint = strings.Contains(msgHost, basehost)
	case "subs":
		shouldPrint = strings.HasSuffix(msgHost, basehost)
	default:
		shouldPrint = true
	}

	if !shouldPrint {
		return false
	}

	c.colorPrint(tag, msg, c.conf.Plain)
	reqsMade.AddReq(getReqFromURL(msg))

	return shouldPrint
}

func (c *Collector) visitWaybackURLs(u string, subdomains *sync.Map, reqsMade *syncList) {
	// get results from waybackurls
	waybackurls := waybackURLs(u)

	// print wayback results, if depth >1, also add them to the crawl queue
	for _, waybackurl := range waybackurls {
		if c.conf.IncludeWayback || c.conf.IncludeAll {
			c.recordIfInScope(c.au.Yellow("[wayback]"), u, waybackurl, reqsMade)
		}
		// if this is a new subdomain, print it
		urlObj, err := url.Parse(waybackurl)
		if err != nil {
			continue
		}
		if c.conf.IncludeSubs || c.conf.IncludeAll {
			if _, exists := subdomains.Load(urlObj.Host); exists {
				continue
			}

			if urlObj.Host != "" && strings.Contains(urlObj.Host, u) {
				c.recordIfInScope(c.au.BrightGreen("[subdomain]"), u, urlObj.Host, reqsMade)
				subdomains.Store(urlObj.Host, struct{}{})
			}
		}
		if c.conf.Depth > 1 {
			c.colly.Visit(waybackurl)
		}
	}
}

var re = regexp.MustCompile(".*llow: ")

func (c *Collector) parseRobots(url string, reqsMade *syncList) {
	var robotsurls []string
	robotsURL := url + "/robots.txt"

	resp, err := http.Get(robotsURL)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	lines := strings.Split(string(body), "\n")

	for _, line := range lines {
		if re.MatchString(line) {
			urlstring := re.ReplaceAllString(line, "")
			if c.conf.IncludeRobots || c.conf.IncludeAll {
				_ = c.recordIfInScope(c.au.BrightMagenta("[robots]"), url, url+urlstring, reqsMade)
			}
			//add it to a slice for parsing later
			robotsurls = append(robotsurls, url+urlstring)
		}
	}

	// if depth is greater than 1, add all of the robots urls as seeds
	if c.conf.Depth > 1 {
		for _, robotsurl := range robotsurls {
			c.colly.Visit(robotsurl)
		}
	}
}

var linkFinderRegex = regexp.MustCompile(`(?:"|')(((?:[a-zA-Z]{1,10}://|//)[^"'/]{1,}\.[a-zA-Z]{2,}[^"']{0,})|((?:/|\.\./|\./)[^"'><,;| *()(%%$^/\\\[\]][^"'><,;|()]{1,})|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{1,}\.(?:[a-zA-Z]{1,4}|action)(?:[\?|/][^"|']{0,}|))|([a-zA-Z0-9_\-]{1,}\.(?:php|asp|aspx|jsp|json|action|html|js|txt|xml)(?:\?[^"|']{0,}|)))(?:"|')`)

func (c *Collector) linkfinder(jsfile string, tag aurora.Value, plain bool) {
	// skip tls verification
	client := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(jsfile)
	if err != nil || resp.StatusCode != 200 {
		return
	}

	// if the js file exists
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	resp.Body.Close()
	found := linkFinderRegex.FindAllString(string(res), -1)
	for _, link := range found {
		c.colorPrint(tag, link + " from " + jsfile, plain)
	}

}

func (c *Collector) parseSitemap(url string, reqsMade *syncList) {
	sitemapURL := url + "/sitemap.xml"
	sitemap.ParseFromSite(sitemapURL, func(e sitemap.Entry) error {
		if c.conf.IncludeSitemap || c.conf.IncludeAll {
			_ = c.recordIfInScope(c.au.BrightBlue("[sitemap]"), url, e.GetLocation(), reqsMade)
		}
		// if depth is greater than 1, add sitemap url as seed
		if c.conf.Depth > 1 {
			c.colly.Visit(e.GetLocation())
		}
		return nil
	})
}

// if -plain is set, just print the message, otherwise print a coloured tag and then the message
func (c *Collector) colorPrint(tag aurora.Value, msg string, plain bool) {
	if plain {
		c.w.Write([]byte(fmt.Sprintln(msg)))
	} else {
		// append message to ansi code as bytes
		bs := append([]byte(fmt.Sprint(tag)), " "...)
		bs = append(bs, []byte(fmt.Sprintln(msg))...)
		c.w.Write(bs)
	}
}

func getReqFromURL(url string) *http.Request {
	// some sanity checking
	if !strings.Contains(url, "http") {
		return nil
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("dello")
		return nil
	}
	return req
}
