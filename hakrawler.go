package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type Result struct {
	Source string
	URL    string
	Where  string
}

var headers map[string]string

func main() {
	inside := flag.Bool("i", false, "Only crawl inside path")
	threads := flag.Int("t", 8, "Number of threads to utilise.")
	depth := flag.Int("d", 2, "Depth to crawl.")
	maxSize := flag.Int("size", -1, "Page size limit, in KB.")
	insecure := flag.Bool("insecure", false, "Disable TLS verification.")
	subsInScope := flag.Bool("subs", false, "Include subdomains for crawling.")
	showJson := flag.Bool("json", false, "Output as JSON.")
	showSource := flag.Bool("s", false, "Show the source of URL based on where it was found. E.g. href, form, script, etc.")
	showWhere := flag.Bool("w", false, "Show at which link the URL is found.")
	rawHeaders := flag.String("h", "", "Custom headers separated by two semi-colons. E.g. -h \"Cookie: foo=bar;;Referer: http://example.com/\" ")
	unique := flag.Bool("u", false, "Show only unique urls.")
	proxy := flag.String("proxy", "", "Proxy URL. E.g. -proxy http://127.0.0.1:8080")
	timeout := flag.Int("timeout", -1, "Maximum time to crawl each URL from stdin, in seconds.")
	disableRedirects := flag.Bool("dr", false, "Disable following HTTP redirects.")

	flag.Parse()

	if *proxy != "" {
		os.Setenv("PROXY", *proxy)
	}
	proxyURL, _ := url.Parse(os.Getenv("PROXY"))

	err := parseHeaders(*rawHeaders)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing headers:", err)
		os.Exit(1)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "No urls detected. Hint: cat urls.txt | hakrawler")
		os.Exit(1)
	}

	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: *insecure},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	regexMu := sync.Mutex{}
	regexCache := make(map[string]*regexp.Regexp)

	done := make(chan struct{})
	inputURLs := make(chan string, *threads*2)
	results := make(chan string, *threads)

	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			inputURLs <- s.Text()
		}
		close(inputURLs)
	}()

	var wg sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rawURL := range inputURLs {
				hostname, err := extractHostname(rawURL)
				if err != nil {
					log.Println("Error parsing URL:", err)
					continue
				}

				allowedDomains := []string{hostname}
				if headers != nil {
					if val, ok := headers["Host"]; ok {
						allowedDomains = append(allowedDomains, val)
					}
				}

				c := colly.NewCollector(
					colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0"),
					colly.Headers(headers),
					colly.AllowedDomains(allowedDomains...),
					colly.MaxDepth(*depth),
					colly.Async(true),
				)

				if *maxSize != -1 {
					c.MaxBodySize = *maxSize * 1024
				}

				if *subsInScope {
					c.AllowedDomains = nil
					regexMu.Lock()
					r, ok := regexCache[hostname]
					if !ok {
						r = regexp.MustCompile(".*(\\.|\\/\\/)" + strings.ReplaceAll(hostname, ".", "\\.") + "((#|\\/|\\?).*)?")
						regexCache[hostname] = r
					}
					regexMu.Unlock()
					c.URLFilters = []*regexp.Regexp{r}
				}

				if *disableRedirects {
					c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					})
				}

				c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: *threads})

				targetURL := rawURL

				c.OnHTML("a[href]", func(e *colly.HTMLElement) {
					link := e.Attr("href")
					absLink := e.Request.AbsoluteURL(link)
					if strings.Contains(absLink, targetURL) || !*inside {
						printResult(link, "href", *showSource, *showWhere, *showJson, results, done, e)
						e.Request.Visit(link)
					}
				})

				c.OnHTML("script[src]", func(e *colly.HTMLElement) {
					printResult(e.Attr("src"), "script", *showSource, *showWhere, *showJson, results, done, e)
				})

				c.OnHTML("form[action]", func(e *colly.HTMLElement) {
					printResult(e.Attr("action"), "form", *showSource, *showWhere, *showJson, results, done, e)
				})

				if headers != nil {
					c.OnRequest(func(r *colly.Request) {
						for header, value := range headers {
							r.Headers.Set(header, value)
						}
					})
				}

				c.WithTransport(&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
					MaxIdleConns: 100,
					MaxIdleConnsPerHost: 10,
					IdleConnTimeout: 90 * time.Second,
				})

				if *timeout == -1 {
					c.Visit(rawURL)
					c.Wait()
				} else {
					finished := make(chan struct{}, 1)
					go func() {
						c.Visit(rawURL)
						c.Wait()
						finished <- struct{}{}
					}()
					select {
					case <-finished:
					case <-time.After(time.Duration(*timeout) * time.Second):
						log.Println("[timeout] " + rawURL)
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(done)
		close(results)
	}()

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	seen := make(map[string]struct{})
	urlsFound := false
	for res := range results {
		if *unique {
			if _, ok := seen[res]; ok {
				continue
			}
			seen[res] = struct{}{}
		}
		fmt.Fprintln(w, res)
		urlsFound = true
	}

	if !urlsFound {
		fmt.Fprintln(os.Stderr, "No URLs were found. This usually happens when a domain is specified (https://example.com), but it redirects to a subdomain (https://www.example.com). The subdomain is not included in the scope, so the no URLs are printed. In order to overcome this, either specify the final URL in the redirect chain or use the -subs option to include subdomains.")
	}
}

// parseHeaders does validation of headers input and saves it to a formatted map.
func parseHeaders(rawHeaders string) error {
	if rawHeaders != "" {
		if !strings.Contains(rawHeaders, ":") {
			return errors.New("headers flag not formatted properly (no colon to separate header and value)")
		}

		headers = make(map[string]string)
		rawHeaders := strings.Split(rawHeaders, ";;")
		for _, header := range rawHeaders {
			var parts []string
			if strings.Contains(header, ": ") {
				parts = strings.SplitN(header, ": ", 2)
			} else if strings.Contains(header, ":") {
				parts = strings.SplitN(header, ":", 2)
			} else {
				continue
			}
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return nil
}

// extractHostname extracts the hostname from a URL and returns it.
func extractHostname(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil || !u.IsAbs() {
		return "", errors.New("Input must be a valid absolute URL")
	}
	return u.Hostname(), nil
}

// printResult constructs output lines and sends them to the results channel.
func printResult(link string, sourceName string, showSource bool, showWhere bool, showJson bool, results chan string, done chan struct{}, e *colly.HTMLElement) {
	result := e.Request.AbsoluteURL(link)
	whereURL := e.Request.URL.String()
	if result != "" {
		if showJson {
			where := ""
			if showWhere {
				where = whereURL
			}
			bytes, _ := json.Marshal(Result{
				Source: sourceName,
				URL:    result,
				Where:  where,
			})
			result = string(bytes)
		} else if showSource {
			result = "[" + sourceName + "] " + result
		}

		if showWhere && !showJson {
			result = "[" + whereURL + "] " + result
		}

		select {
		case results <- result:
		case <-done:
		}
	}
}
