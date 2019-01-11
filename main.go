package main

import (
	"fmt" //GO’s base package
	"io"
	"net/http"    //for sending HTTP requests
	"net/url"     //for URL formatting
	"strings"     //string manipulation and testing
	"sync"        //for thread safe map
	"sync/atomic" //for thread safe map
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/romana/rlog"
)

//URL filter function definition
type filterFunc func(string, Crawler) bool

//Our crawler structure definition
type Crawler struct {
	//the base URL of the website being crawled
	host string
	//a channel on which the crawler will receive new (unfiltered) URLs to crawl
	//the crawler will pass everything received from this channel
	//through the chain of filters we have
	//and only allowed URLs will be passed to the filteredUrls channel
	urls chan string
	//a channel on which the crawler will receive filtered URLs.
	filteredUrls chan string
	// channel on which we will be sending quit signal
	quit chan string
	//a slice that contains the filters we want to apply on the URLs.
	filters []filterFunc
	// Depth of links to visit
	depth sync.Map
	// Visited urls
	visited sync.Map
	// Politeness delay for crawler in seconds
	politeness int
	//an integer to track how many URLs have been crawled
	count int
	// count of threads in processing
	processing int32
	// channel for adding webpages
	webpages chan Webpage
}

//starts the crawler
//the method starts two GO functions
//the first one waits for new URLs as they
//get extracted.
//the second waits for filtered URLs as they
//pass through all the registered filters
func (crawler *Crawler) start(wsite *website) {
	//wait for new URLs to be extracted and passed to the URLs channel.
	go func() {
		for {
			select {
			case url := <-crawler.urls:
				atomic.AddInt32(&crawler.processing, 1)
				go crawler.filter(url)
			case <-crawler.quit:
				log.Debugf("> Closing urls channel")
				close(crawler.urls)
				return
			}
		}
	}()

	//wait for filtered URLs to arrive through the filteredUrls channel
	go func() {
		for {
			select {
			case url := <-crawler.filteredUrls:
				crawler.count++
				// log.Debugf("%d: Crawling %s ", crawler.count, url)
				atomic.AddInt32(&crawler.processing, 1)
				wpage := Webpage{url, nil}
				go crawler.crawl(&wpage)
				log.Infof("Waiting for %d second before next requests", crawler.politeness)
				time.Sleep(time.Duration(crawler.politeness) * time.Second)
			case <-crawler.quit:
				log.Debugf("> Closing filteredUrls channel")
				close(crawler.filteredUrls)
				return
			}
		}
	}()

	// Collect website objects from crawling
	go func() {
		for {
			select {
			case wpage := <-crawler.webpages:
				wsite.AddWebpage(wpage)
				log.Infof("Added %s", wpage.URL)
			case <-crawler.quit:
				close(crawler.webpages)
				return
			}
		}
	}()
}

//given a URL, the method will apply all the filters
//on that URL, if and only if, it passes through all
//the filters, it will then be passed to the filteredUrls channel
func (crawler *Crawler) filter(url string) {
	defer func() { atomic.AddInt32(&crawler.processing, -1) }()
	temp := false
	for _, fn := range crawler.filters {
		temp = fn(url, *crawler)
		if temp != true {
			return
		}
	}
	atomic.AddInt32(&crawler.processing, 1)
	go func() {
		defer func() { atomic.AddInt32(&crawler.processing, -1) }()
		crawler.filteredUrls <- url
	}()
}

//given a URL, the method will send an HTTP GET request
//extract the response body
//extract the URLs from the body
func (crawler *Crawler) crawl(wpage *Webpage) {
	defer func() { crawler.processing += -1 }()
	url := wpage.URL

	//send http request
	depth, _ := crawler.depth.Load(url)
	visited, _ := crawler.visited.Load(url)
	if !visited.(bool) && depth.(int) <= 2 {
		//here we make call to url
		resp, err := http.Get(url)
		if err != nil {
			log.Debug("An error has occured")
			log.Debug(err)
		} else {
			defer resp.Body.Close()
			if err != nil {
				log.Debug("Error while fetching body for : " + url)
				log.Debug(err)
			} else {
				crawler.extractUrls(wpage, resp.Body)
				log.Debugf("References for %s are %d", wpage.URL, len(wpage.References))
				crawler.visited.Store(url, true)
				crawler.webpages <- *wpage
			}
		}
	} else {
		log.Debugf("For %s:  Depth : %d and visited : %t", url, depth, visited)
	}
	return
}

func (crawler *Crawler) extractUrls(wpage *Webpage, body io.ReadCloser) {
	doc, err := goquery.NewDocumentFromReader(body)
	Url := wpage.URL

	baseURL, _ := url.Parse(Url)
	if err != nil {
		log.Debugf("Error parsing goquery: %s", Url)
		log.Debug(err)
	}

	doc.Find("body a").Each(func(i int, s *goquery.Selection) {
		raw_href, ok := s.Attr("href")
		if ok {
			href, _ := url.Parse(raw_href)
			// Resolve the relative urls
			if strings.HasPrefix(raw_href, "/") ||
				strings.HasPrefix(raw_href, ".") {
				href = baseURL.ResolveReference(href)
			}

			wpage.References = append(wpage.References, href.String())
			_, visited := crawler.visited.Load(href.String())

			if !visited {
				crawler.visited.Store(href.String(), false)
				depth, _ := crawler.depth.Load(Url)
				crawler.depth.Store(href.String(), depth.(int)+1)
				atomic.AddInt32(&crawler.processing, 1)
				go func() {
					defer func() { atomic.AddInt32(&crawler.processing, -1) }()
					crawler.urls <- href.String()
				}()
			}
		}
	})
}

//adds a new URL filter to the crawler
func (crawler *Crawler) addFilter(filter filterFunc) Crawler {
	crawler.filters = append(crawler.filters, filter)
	return *crawler
}

func main() {
	//create a new instance of the crawler structure
	// startURL := "https://sumit.murari.me"
	startURL := "https://monzo.com/"
	wsite := website{startURL, nil}
	c := Crawler{
		startURL,
		make(chan string, 10),
		make(chan string, 10),
		make(chan string),
		make([]filterFunc, 0),
		sync.Map{},
		sync.Map{},
		2,
		0,
		0,
		make(chan Webpage, 10),
	}

	c.addFilter(IsInternal)
	c.addFilter(IsValidPath)
	c.addFilter(IsValidSubdomain)

	c.depth.Store(startURL, 0)
	c.visited.Store(startURL, false)

	c.start(&wsite)
	c.urls <- c.host
	for {
		log.Debugf("urls queue: %d ; filteredUrls queue: %d; processing: %d ", len(c.urls), len(c.filteredUrls), c.processing)
		time.Sleep(2 * time.Second)

		if len(c.filteredUrls) == 0 && len(c.urls) == 0 && c.processing == 0 {
			log.Debugf("urls and filteredUrls channels and no url in processing")
			c.quit <- "done"
			c.quit <- "done"
			c.quit <- "done"
			break
		}
	}
	fmt.Println("Printing website")
	wsite.PrintBasicSiteMap()
	wsite.PrintSiteGraph()
	fmt.Println("Good bye")
}
