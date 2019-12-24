# hakrawler
Simple web crawler written in Golang

## Usage
```
$ hakrawler -h
Usage of hakrawler:
  -depth int
    	Maximum depth to crawl (default 1)
  -domain string
    	Domain to crawl (default "http://127.0.0.1")
  -js
    	Include links to utilised JavaScript files and exclude all other details
```

## Example
```
$ hakrawler -domain google.com -depth 1
http://www.google.com.au/imghp?hl=en&tab=wi
http://maps.google.com.au/maps?hl=en&tab=wl
https://play.google.com/?hl=en&tab=w8
http://www.youtube.com/?gl=AU&tab=w1
http://news.google.com.au/nwshp?hl=en&tab=wn
https://mail.google.com/mail/?tab=wm
https://drive.google.com/?tab=wo
https://www.google.com.au/intl/en/about/products?tab=wh
http://www.google.com.au/history/optout?hl=en
http://www.google.com/preferences?hl=en
https://accounts.google.com/ServiceLogin?hl=en&passive=true&continue=http://www.google.com/
http://www.google.com/advanced_search?hl=en-AU&authuser=0
http://www.google.com/language_tools?hl=en-AU&authuser=0
http://www.google.com/intl/en/ads/
http://www.google.com/services/
http://www.google.com/intl/en/about.html
http://www.google.com/setprefdomain?prefdom=AU&prev=http://www.google.com.au/&sig=K_m93KYiK2mCsMUgBgjUxDQO9XjKU%3D
http://www.google.com/intl/en/policies/privacy/
http://www.google.com/intl/en/policies/terms/
https://www.google.com.au/imghp?hl=en&tab=wi
https://maps.google.com.au/maps?hl=en&tab=wl
https://www.youtube.com/?gl=AU&tab=w1
https://news.google.com.au/nwshp?hl=en&tab=wn
https://www.google.com/preferences?hl=en
https://accounts.google.com/ServiceLogin?hl=en&passive=true&continue=https://www.google.com/
https://www.google.com/advanced_search?hl=en-AU&authuser=0
https://www.google.com/language_tools?hl=en-AU&authuser=0
https://www.google.com/intl/en/ads/
https://www.google.com/services/
https://www.google.com/intl/en/about.html
https://www.google.com/setprefdomain?prefdom=AU&prev=https://www.google.com.au/&sig=K_N-GAtzESSj98REoQykMbaSB_bUQ%3D
https://www.google.com/intl/en/policies/privacy/
https://www.google.com/intl/en/policies/terms/
```
