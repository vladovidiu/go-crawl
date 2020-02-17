package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/steelx/extractlinks"
)

var (
	config = &tls.Config{
		InsecureSkipVerify: true,
	}
	transport = &http.Transport{
		TLSClientConfig: config,
	}
	netClient = &http.Client{
		Transport: transport,
	}
	queue      = make(chan string)
	hasVisited = make(map[string]bool)
)

func main() {
	arguments := os.Args[1:]

	if len(arguments) == 0 {
		fmt.Println("Missing URL")
		os.Exit(1)
	}

	baseURL := arguments[0]

	go func() {
		queue <- baseURL
	}()

	for href := range queue {
		if !hasVisited[href] && isSameDomain(href, baseURL) {
			crawlUrl(href)
		}
	}
}

func crawlUrl(href string) {
	hasVisited[href] = true
	fmt.Printf("Crawling url -> %v \n", href)
	response, err := netClient.Get(href)
	defer response.Body.Close()
	checkErr(err)

	links, err := extractlinks.All(response.Body)
	checkErr(err)

	for _, link := range links {
		go func() {
			queue <- toFixedUrl(link.Href, href)
		}()
	}
}

func toFixedUrl(href, baseUrl string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}

	base, err := url.Parse(baseUrl)
	if err != nil {
		return ""
	}

	toFixedUri := base.ResolveReference(uri)

	return toFixedUri.String()
}

func isSameDomain(href, baseUrl string) bool {
	uri, err := url.Parse(href)
	if err != nil {
		return false
	}

	parentUri, err := url.Parse(baseUrl)
	if err != nil {
		return false
	}

	if uri.Host != parentUri.Host {
		return false
	}

	return true
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
