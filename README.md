# Hakrawler

Fast golang web crawler for gathering URLs and JavaScript file locations. This is basically a simple implementation of the awesome Gocolly library.

## Example usages

Single URL:

```
echo https://google.com | hakrawler
```

Multiple URLs:

```
cat urls.txt | hakrawler
```

Timeout for each line of stdin after 5 seconds:

```
cat urls.txt | hakrawler -timeout 5
```

Send all requests through a proxy:

```
cat urls.txt | hakrawler -proxy http://localhost:8080
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

### Normal Install

First, you'll need to [install go](https://golang.org/doc/install).

Then run this command to download + compile hakrawler:
```
go install github.com/hakluke/hakrawler@latest
```

You can now run `~/go/bin/hakrawler`. If you'd like to just run `hakrawler` without the full path, you'll need to `export PATH="~/go/bin/:$PATH"`. You can also add this line to your `~/.bashrc` file if you'd like this to persist.

### Docker Install (from dockerhub)

```
echo https://www.google.com | docker run --rm -i hakluke/hakrawler:v2 -subs
```

### Local Docker Install

It's much easier to use the dockerhub method above, but if you'd prefer to run it locally:

```
git clone https://github.com/hakluke/hakrawler
cd hakrawler
sudo docker build -t hakluke/hakrawler .
sudo docker run --rm -i hakluke/hakrawler --help
```
### Kali Linux: Using apt

Note: This will install an older version of hakrawler without all the features, and it may be buggy. I recommend using one of the other methods.

```sh
sudo apt install hakrawler
```

Then, to run hakrawler:

```
echo https://www.google.com | docker run --rm -i hakluke/hakrawler -subs
```

## Command-line options
```
Usage of hakrawler:
  -d int
    	Depth to crawl. (default 2)
  -dr
    	Disable following HTTP redirects.
  -h string
    	Custom headers separated by two semi-colons. E.g. -h "Cookie: foo=bar;;Referer: http://example.com/"
  -i	Only crawl inside path
  -insecure
    	Disable TLS verification.
  -json
    	Output as JSON.
  -proxy string
    	Proxy URL. E.g. -proxy http://127.0.0.1:8080
  -s	Show the source of URL based on where it was found. E.g. href, form, script, etc.
  -size int
    	Page size limit, in KB. (default -1)
  -subs
    	Include subdomains for crawling.
  -t int
    	Number of threads to utilise. (default 8)
  -timeout int
    	Maximum time to crawl each URL from stdin, in seconds. (default -1)
  -u	Show only unique urls.
  -w	Show at which link the URL is found.
```
