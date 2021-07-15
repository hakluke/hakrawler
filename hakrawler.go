package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

var headers map[string]string

func main() {
	threads := flag.Int("t", 8, "Number of threads to utilise.")
	depth := flag.Int("d", 2, "Depth to crawl.")
	insecure := flag.Bool("insecure", false, "Disable TLS verification.")
	rawHeaders := flag.String(("h"), "", "Custom headers separated by semi-colon. E.g. -h \"Cookie: foo=bar\" ")
	flag.Parse()

	// Convert the headers input to a usable map (or die trying)
	err := parseHeaders(*rawHeaders)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing headers:", err)
		os.Exit(1)
	}

	// Check for stdin input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "No urls detected. Hint: cat urls.txt | hakrawler")
		os.Exit(1)
	}

	results := make(chan string, *threads)
	go func() {
		// get each line of stdin, push it to the work channel
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			crawl(results, s.Text(), *threads, *depth, *insecure)
		}
		if err := s.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
		close(results)
	}()

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for res := range results {
		fmt.Fprintln(w, res)
	}
}

func crawl(results chan<- string, url string, threads int, depth int, insecure bool) {
	hostname, err := extractHostname(url)
	if err != nil {
		log.Println("Error parsing URL:", err)
		return
	}
	// Instantiate default collector
	c := colly.NewCollector(
		// limit crawling to the domain of the specified URL
		colly.AllowedDomains(hostname),
		// set MaxDepth to the specified depth
		colly.MaxDepth(depth),
		// specify Async for threading
		colly.Async(true),
	)

	// Set parallelism
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: threads})

	// Print every href found, and visit it
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		absoluteURL := e.Request.AbsoluteURL(link)

		if absoluteURL != "" {
			results <- "[href] " + e.Request.AbsoluteURL(link)
			// Visit link found on page on a new thread
			e.Request.Visit(link)
		}
	})

	// find and print all the JavaScript files
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		absoluteURL := e.Request.AbsoluteURL(link)
		if absoluteURL != "" {
			results <- "[script] " + e.Request.AbsoluteURL(link)
		}
	})

	// find and print all the JavaScript files
	c.OnHTML("form[action]", func(e *colly.HTMLElement) {
		link := e.Attr("action")
		absoluteURL := e.Request.AbsoluteURL(link)
		if absoluteURL != "" {
			results <- "[form] " + e.Request.AbsoluteURL(link)
		}
	})

	// add the custom headers
	if headers != nil {
		c.OnRequest(func(r *colly.Request) {
			for header, value := range headers {
				r.Headers.Set(header, value)
			}
		})
	}

	// Skip TLS verification if -insecure flag is present
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	})

	// Start scraping
	c.Visit(url)
	// Wait until threads are finished
	c.Wait()
}

// parseHeaders does validation of headers input and saves it to a formatted map.
func parseHeaders(rawHeaders string) error {
	if rawHeaders != "" {
		if !strings.Contains(rawHeaders, ":") {
			return errors.New("headers flag not formatted properly (no colon to separate header and value)")
		}

		headers = make(map[string]string)
		rawHeaders := strings.Split(rawHeaders, ";")
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
	}
	return nil
}

// extractHostname() extracts the hostname from a URL and returns it
func extractHostname(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}
