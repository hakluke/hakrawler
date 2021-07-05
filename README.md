# Hakrawler

Fast golang web crawler for gathering URLs and JavaSript file locations.

## Example usage

```
cat urls.txt | hakrawler
```

## Example tool chain

Get all subdomains of google, find the ones that respond to http(s), crawl them all.

```
echo google.com | haktrails subdomains | httpx | hakrawler
```

## Usage gif

![Example usage gif](hakrawler-example.gif)

## Command-line options
```
  -d int
    	Depth to crawl. (default 2)
  -insecure
    	Disable TLS verification.
  -t int
    	Number of threads to utilise. (default 8)
```

## Version 2 note

From version 2, hakrawler has been completely rewritten and dramatically simplified to align more closely with the unix philosophy.

- It is now much faster and less buggy.
- Many features have been deprecated (robots.txt parsing, JS file parsing, sitemap parsing, waybackurls), instead, these features are written into separate tools that can be piped to from hakrawler.
- No more terminal colours because they can cause annoying issues when piping to other tools.
- Version 1 was my first ever Go project and the code was bad.
