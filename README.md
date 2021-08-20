# Hakrawler

Fast golang web crawler for gathering URLs and JavaSript file locations. This is basically a simple implementation of the awesome Gocolly library.

## Example usages

Single URL:

```
echo https://google.com | hakrawler
```

Multiple URLs:

```
cat urls.txt | hakrawler
```

Include subdomains:

```
echo https://google.com | hakrawler -subs
```

> Note: a common issue is that the tool returns no URLs. This usually happens when a domain is specified (https://example.com), but it redirects to a subdomain (https://www.example.com). The subdomain is not included in the scope, so the no URLs are printed. In order to overcome this, either specify the final URL in the redirect chain or use the `-subs` option to include subdomains.

## Example tool chain

Get all subdomains of google, find the ones that respond to http(s), crawl them all.

```
echo google.com | haktrails subdomains | httpx | hakrawler
```

## Installation

First, you'll need to [install go](https://golang.org/doc/install).

Then run this command to download + compile hakrawler:
```
go install github.com/hakluke/hakrawler@latest
```

You can now run `~/go/bin/hakrawler`. If you'd like to just run `hakrawler` without the full path, you'll need to `export PATH="/go/bin/:$PATH"`. You can also add this line to your `~/.bashrc` file if you'd like this to persist.

## Command-line options
```
  -d int
    	Depth to crawl. (default 2)
  -h string
    	Custom headers separated by two semi-colons. E.g. -h "Cookie: foo=bar;;Referer: http://example.com/"
  -insecure
    	Disable TLS verification.
  -s	Show the source of URL based on where it was found (href, form, script, etc.)
  -subs
    	Include subdomains for crawling.
  -t int
    	Number of threads to utilise. (default 8)
  -u	Show only unique urls
```

## Version 2 note

From version 2, hakrawler has been completely rewritten and dramatically simplified to align more closely with the unix philosophy.

- It is now much faster and less buggy.
- Many features have been deprecated (robots.txt parsing, JS file parsing, sitemap parsing, waybackurls), instead, these features are written into separate tools that can be piped to from hakrawler.
- No more terminal colours because they can cause annoying issues when piping to other tools.
- Version 1 was my first ever Go project and the code was bad.
