package main

import (
	// "fmt"
	"io"
	"os"
	"strings"

	log "github.com/romana/rlog"
)

// A webpage is url with refrences to other webpages
type Webpage struct {
	URL        string
	References []string
}

// A website is a list of webpages
type website struct {
	name     string
	webpages []Webpage
}

// method
func (w *website) AddWebpage(wpage Webpage) {
	w.webpages = append(w.webpages, wpage)
}

func (w *website) PrintBasicSiteMap() {
	log.Debug("==== Printing Sitemap ====")
	log.Infof("Number of webpages: %d ", len(w.webpages))
	log.Info("Sitemap at : sitemap.txt")

	file, err := os.Create("sitemap.txt")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	for _, page := range w.webpages {
		io.Copy(file, strings.NewReader(page.URL+"\n"))
		// io.Copy(file, strings.NewReader(page.String()))
	}

	log.Info("==== Sitemap Done  ==== ")
}

func (w *website) PrintSiteGraph() {
	log.Debug("====  Printing SiteGraph ====")
	log.Info("Sitegraph at : \"sitegraph.txt\"")

	file, err := os.Create("sitegraph.txt")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	for _, page := range w.webpages {
		// fmt.Fprintln(file, page.String())
		io.Copy(file, strings.NewReader(page.URL+"\n"))
		for _, reference := range page.References {
			io.Copy(file, strings.NewReader("-> "+reference+"\n"))
			// fmt.Fprintln(file, "-> "+reference.String())
		}
	}
	log.Info("==== SiteGraph done ==== ")
}

func CreateWebSite(siteurl string) *website {
	wsite := website{siteurl, nil}
	return &wsite
}
