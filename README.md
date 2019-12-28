# hakrawler

[![Twitter](https://img.shields.io/badge/twitter-@hakluke_-blue.svg)](https://twitter.com/hakluke)
[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/hakluke/hakrawler)
[![License](https://img.shields.io/badge/license-GPL3-_red.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html)  

## What is it?

hakrawler is a Go web crawler designed for easy, quick discovery of endpoints and assets within a web application. It can be used to discover:

- Forms
- Endpoints
- Subdomains
- Related domains
- JavaScript files

The goal is to create the tool in a way that it can be easily chained with other tools such as subdomain enumeration tools and vulnerability scanners in order to facilitate tool chaining, for example:

```
amass | hakrawler | some-xss-scanner
```

## Features

- Unlimited, fast web crawling for endpoint discovery
- Fuzzy matching for subdomain discovery
- robots.txt parsing
- sitemap.xml parsing
- Plain output for easy parsing into other tools
- Accept domains from stdin for easier tool chaining

## Upcoming features

- SQLMap-friendly output format
- Link gathering from JavaScript files

## Credits

- [codingo](https://twitter.com/codingo_) and [prodigysml/sml555](https://twitter.com/sml555_), my favourite people to hack with. A constant source of ideas and inspiration. They also provided beta testing and a sounding board for this tool in development.
- [tomnomnom](https://twitter.com/tomnomnom) who wrote waybackurls, which powers the wayback part of this tool
- [s0md3v](https://twitter.com/s0md3v) who wrote photon, which I took ideas from to create this tool
- The folks from [gocolly](https://github.com/gocolly/colly), the library which powers the crawler engine
- [oxffaa](https://github.com/oxffaa/), who wrote a very efficient sitemap.xml parser which is used in this tool

## Installation
1. Install Golang
2. Run the command below
```
go get github.com/hakluke/hakrawler
```
3. Run hakrawler from your Go bin directory. For linux systems it will likely be:
```
~/go/bin/hakrawler
```
Note that if you need to do this, you probably want to add your Go bin directory to your $PATH to make things easier!

## Usage
Note: multiple domains can be crawled by piping them into hakrawler from stdin. If only a single domain is being crawled, it can be added by using the -domain flag.
```
  -h
      Show usage
  -domain string
      The domain that you wish to crawl (for example, google.com)
  -all
    	Include all sources in output (default true)
      This is an alias for including all of the following options: -subs, -urls, -sitemap, -robots, -js and -forms
  -scope string
    	Scope to include:
    	strict = specified domain only
    	subs = specified domain and subdomains
    	fuzzy = any results where the domain contains the supplied domain (for example URLs from withgoogle.com would be returned if you supply the domain google.com)
    	yolo = everything (default "subs")
  -depth int
      Maximum depth to crawl, the default is 1. Anything above 1 will include URLs from robots, sitemap, waybackurls and the initial crawler as a seed. Higher numbers take longer but yield more results.
  -plain
    	Don't use colours or print the banner to allow for easier parsing 
  -schema string
    	Schema, http or https (default "http")
  -usewayback
    	Query wayback machine for URLs and add them as seeds for the crawler
  -cookie
      The value passed to this will be used as the "Cookie" header on the crawler requests
  -auth
      The value passed to this will be used as the "Authorization" header on the crawler requests 
  -forms
    	Include form actions in output
  -js
    	Include links to utilised JavaScript files in output
  -robots
    	Include robots.txt entries in output
  -sitemap
    	Include sitemap.xml entries in output
  -subs
    	Include subdomains in output
  -urls
    	Include URLs discovered by the crawler in output
  -wayback
    	Include wayback machine entries in output
```

## Basic Example

### Image:

Command: `hakrawler -domain bugcrowd.com -depth 1`

![sample output](./hakrawler-output-sample.png)

### Full text output:

```
   $ hakrawler -domain bugcrowd.com -depth 1

██╗  ██╗ █████╗ ██╗  ██╗██████╗  █████╗ ██╗    ██╗██╗     ███████╗██████╗
██║  ██║██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██║    ██║██║     ██╔════╝██╔══██╗
███████║███████║█████╔╝ ██████╔╝███████║██║ █╗ ██║██║     █████╗  ██████╔╝
██╔══██║██╔══██║██╔═██╗ ██╔══██╗██╔══██║██║███╗██║██║     ██╔══╝  ██╔══██╗
██║  ██║██║  ██║██║  ██╗██║  ██║██║  ██║╚███╔███╔╝███████╗███████╗██║  ██║
╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚══╝╚══╝ ╚══════╝╚══════╝╚═╝  ╚═╝
                        Crafted with <3 by hakluke
[robots] http://bugcrowd.com/*?preview
[sitemap] https://bugcrowd.com/
[sitemap] https://bugcrowd.com/contact/
[sitemap] https://bugcrowd.com/faq/
[sitemap] https://bugcrowd.com/leaderboard/
[sitemap] https://bugcrowd.com/list-of-bug-bounty-programs/
[sitemap] https://bugcrowd.com/press/
[sitemap] https://bugcrowd.com/pricing/
[sitemap] https://bugcrowd.com/privacy/
[sitemap] https://bugcrowd.com/terms/
[sitemap] https://bugcrowd.com/resources/responsible-disclosure-program/
[sitemap] https://bugcrowd.com/resources/why-care-about-web-security/
[sitemap] https://bugcrowd.com/resources/what-is-a-bug-bounty/
[sitemap] https://bugcrowd.com/stories/movember/
[sitemap] https://bugcrowd.com/stories/riskio/
[sitemap] https://bugcrowd.com/stories/tagged/
[sitemap] https://bugcrowd.com/tour/
[sitemap] https://bugcrowd.com/tour/platform/
[sitemap] https://bugcrowd.com/tour/crowd/
[sitemap] https://bugcrowd.com/customers/programs/new
[sitemap] https://bugcrowd.com/portal/
[sitemap] https://bugcrowd.com/portal/user/sign_in/
[sitemap] https://bugcrowd.com/portal/user/sign_up/
[url] https://bugcrowd.com/user/sign_in
[subdomain] bugcrowd.com
[url] https://tracker.bugcrowd.com/user/sign_in
[subdomain] tracker.bugcrowd.com
[url] https://www.bugcrowd.com/
[subdomain] www.bugcrowd.com
[url] https://www.bugcrowd.com/products/how-it-works/
[url] https://www.bugcrowd.com/products/how-it-works/the-bugcrowd-difference/
[url] https://www.bugcrowd.com/products/platform/
[url] https://www.bugcrowd.com/products/platform/integrations/
[url] https://www.bugcrowd.com/products/platform/vulnerability-rating-taxonomy/
[url] https://www.bugcrowd.com/products/attack-surface-management/
[url] https://www.bugcrowd.com/products/bug-bounty/
[url] https://www.bugcrowd.com/products/vulnerability-disclosure/
[url] https://www.bugcrowd.com/products/next-gen-pen-test/
[url] https://www.bugcrowd.com/products/bug-bash/
[url] https://www.bugcrowd.com/resources/reports/priority-one-report
[url] https://www.bugcrowd.com/solutions/
[url] https://www.bugcrowd.com/solutions/financial-services/
[url] https://www.bugcrowd.com/solutions/healthcare/
[url] https://www.bugcrowd.com/solutions/retail/
[url] https://www.bugcrowd.com/solutions/automotive-security/
[url] https://www.bugcrowd.com/solutions/technology/
[url] https://www.bugcrowd.com/solutions/government/
[url] https://www.bugcrowd.com/solutions/security/
[url] https://www.bugcrowd.com/solutions/marketplace-apps/
[url] https://www.bugcrowd.com/customers/
[url] https://www.bugcrowd.com/hackers/
[url] https://bugcrowd.com/programs
[url] https://bugcrowd.com/crowdstream
[url] https://www.bugcrowd.com/bug-bounty-list/
[url] https://www.bugcrowd.com/hackers/faqs/
[url] https://www.bugcrowd.com/resources/help-wanted/
[url] https://www.bugcrowd.com/hackers/bugcrowd-university/
[url] https://www.bugcrowd.com/hackers/ambassador-program/
[url] https://forum.bugcrowd.com
[subdomain] forum.bugcrowd.com
[url] https://bugcrowd.com/leaderboard
[url] https://www.bugcrowd.com/resources/levelup-0x04
[url] https://www.bugcrowd.com/resources/
[url] https://www.bugcrowd.com/resources/webinars/
[url] https://www.bugcrowd.com/resources/bakers-dozen/
[url] https://www.bugcrowd.com/events/
[url] https://www.bugcrowd.com/resources/glossary/
[url] https://www.bugcrowd.com/resources/faqs/
[url] https://www.bugcrowd.com/about/
[url] https://www.bugcrowd.com/blog
[url] https://www.bugcrowd.com/about/expertise/
[url] https://www.bugcrowd.com/about/leadership/
[url] https://www.bugcrowd.com/about/press-releases/
[url] https://www.bugcrowd.com/about/careers/
[url] https://www.bugcrowd.com/partners/
[url] https://www.bugcrowd.com/about/news/
[url] https://www.bugcrowd.com/about/contact/
[url] https://bugcrowd.com/user/sign_up
[url] https://www.bugcrowd.com/get-started/
[url] https://www.bugcrowd.com/products/attack-surface-management
[url] https://www.bugcrowd.com/products/bug-bounty
[url] https://www.bugcrowd.com/customers/motorola
[url] https://www.bugcrowd.com/products/vulnerability-disclosure
[url] https://www.bugcrowd.com/products/next-gen-pen-test
[url] https://www.bugcrowd.com/resources/guides/esg-research-ciso-security-trends
[url] https://www.bugcrowd.com/events/join-us-at-rsa-2019-march-4-8-2019-san-francisco/
[url] https://www.bugcrowd.com/resources/4-reasons-to-swap-your-traditional-pen-test-with-a-next-gen-pen-test/
[url] https://www.bugcrowd.com/blog/november-2019-hall-of-fame/
[url] https://www.bugcrowd.com/blog/bugcrowd-launches-crowdstream-and-in-platform-coordinated-disclosure/
[url] https://www.bugcrowd.com/blog/the-future-is-now-2020-cybersecurity-predictions/
[url] https://www.bugcrowd.com/press-release/bugcrowd-launches-first-crowd-driven-approach-to-risk-based-asset-discovery-and-prioritization/
[url] https://www.bugcrowd.com/press-release/bugcrowd-university-expands-education-and-training-for-whitehat-hackers/
[url] https://www.bugcrowd.com/press-release/bugcrowd-announces-industrys-first-platform-enabled-cybersecurity-assessments-for-marketplaces/
[url] https://www.bugcrowd.com/news/
[url] https://www.bugcrowd.com/events/appsec-cali/
[url] https://www.bugcrowd.com/events
[url] https://www.bugcrowd.com/bugcrowd-security/
[url] https://www.bugcrowd.com/terms-and-conditions/
[url] https://www.bugcrowd.com/privacy/
[javascript] https://www.bugcrowd.com/wp-content/uploads/autoptimize/js/autoptimize_single_de6b8fb8b3b0a0ac96d1476a6ef0d147.js
[javascript] https://www.bugcrowd.com/wp-content/uploads/autoptimize/js/autoptimize_79a2bb0d9a869da52bd3e98a65b0cfb7.js

```
