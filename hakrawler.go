package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gocolly/colly"
)

func main() {
	threads := flag.Int("t", 8, "Number of threads to utilise.")
	depth := flag.Int("d", 2, "Depth to crawl.")
	insecure := flag.Bool("insecure", false, "Disable TLS verification.")
	flag.Parse()

	// Check for stdin input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "No urls detected. Hint: cat urls.txt | hakrawler")
		os.Exit(1)
	}

	// get each line of stdin, push it to the work channel
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		crawl(s.Text(), *threads, *depth, *insecure)
	}
	if err := s.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func crawl(url string, threads int, depth int, insecure bool) {
	// Instantiate default collector
	c := colly.NewCollector(
		// MaxDepth is 2, so only the links on the scraped page
		// and links on those pages are visited
		colly.MaxDepth(2),
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
			fmt.Printf("[href] %s\n", e.Request.AbsoluteURL(link))
			// Visit link found on page on a new thread
			e.Request.Visit(link)
		}
	})

	// find and print all the JavaScript files
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		absoluteURL := e.Request.AbsoluteURL(link)
		if absoluteURL != "" {
			fmt.Printf("[script] %s\n", e.Request.AbsoluteURL(link))
		}
	})

	// find and print all the JavaScript files
	c.OnHTML("form[action]", func(e *colly.HTMLElement) {
		link := e.Attr("action")
		absoluteURL := e.Request.AbsoluteURL(link)
		if absoluteURL != "" {
			fmt.Printf("[form] %s\n", e.Request.AbsoluteURL(link))
		}
	})

	// Skip TLS verification if -insecure flag is present
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	})

	// Start scraping
	c.Visit(url)
	// Wait until threads are finished
	c.Wait()
}
