package main

import (
	"net/url"
	"strings"
	// log "github.com/romana/rlog"
)

// Filters
var excludedPath = []string{
	"/cdn-cgi",
	"/legal",
	"/static",
	"/blog",
}
var excludedSubdomain = []string{
	"www.monzo.com",
	"community.monzo.com",
	"status.monzo.com",
}

// Check if internal url
func IsInternal(URL string, crawler Crawler) bool {
	href, _ := url.Parse(URL)
	baseURL, _ := url.Parse(crawler.host)
	if strings.HasSuffix(href.Hostname(), baseURL.Hostname()) {
		return true
	}
	// log.Debug(baseURL.Hostname() + " : doesn't match with : " + href.Hostname())
	return false
}

// Check if request comes is excluded path
func IsValidPath(URL string, crawler Crawler) bool {
	href, _ := url.Parse(URL)
	path := href.Path
	for _, v := range excludedPath {
		if strings.HasPrefix(path, v) {
			// log.Debug(href.Path + " : prefix is in excluded list")
			return false
		}
	}
	return true
}

func IsValidSubdomain(URL string, crawler Crawler) bool {
	href, _ := url.Parse(URL)
	domain := href.Hostname()
	for _, v := range excludedSubdomain {
		if strings.Contains(domain, v) {
			// log.Debug(domain + ": is part of excluded list")
			return false
		}
	}
	return true
}
