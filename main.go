package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cpaules/go-link-parser"
)

/*
Steps:
1. GET the web page
2. parse all links in the web page
3. build proper urls with the links
4. filter out links with a different domain
5. find all pages (BFS)
Repeat steps 1-5 for each page
6. print out XML
*/

//	links that begin with:
//		"/" are assumed to be from the same domain and are thus kept
// 		"https://*urlFlag" are also assumed to be from the same domain
//	any other prefix will be discarded

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func main() {
	urlFlag := flag.String("url", "https://google.com", "the url you want to build a sitemap for")
	maxDepth := flag.Int("depth", 2, "the maxmimum number of links deep to traverse")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)
	toXML := urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		// fmt.Println(page)
		toXML.Urls = append(toXML.Urls, loc{page})
	}

	fmt.Println(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	if err := enc.Encode(toXML); err != nil {
		panic(err)
	}
}

func bfs(urlString string, maxDepth int) []string {
	seen := make(map[string]struct{}) //struct uses less memory than usig bool
	var q map[string]struct{}         // q = current level in tree
	nq := map[string]struct{}{        // all the children of q, a map of urls not yet visited
		urlString: struct{}{},
	}
	// visit each url in q, add all of its children to nq, repeat for all urls in q
	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		if len(q) == 0 {
			break
		}
		for url := range q {
			if _, ok := seen[url]; ok { // if url has been seen
				continue
			}
			seen[url] = struct{}{} // mark url as seen if it passes the if statement
			for _, link := range get(url) {
				if _, ok := seen[link]; !ok {
					nq[link] = struct{}{}
				}
			}
		}
	}
	// var ret []string
	// preallocates memory for slice, so that appending doesnt need to reallocate memory
	ret := make([]string, 0, len(seen))
	for url := range seen {
		ret = append(ret, url)
	}
	return ret
}

func get(urlStr string) []string {
	resp, err := http.Get(urlStr)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// fmt.Println(resp)
	reqURL := resp.Request.URL
	baseURL := &url.URL{
		Scheme: reqURL.Scheme,
		Host:   reqURL.Host,
	}
	base := baseURL.String()
	// fmt.Println(base)
	// pages := hrefs(resp.Body, base)
	return filter(hrefs(resp.Body, base), withPrefix(base))

}

func hrefs(r io.Reader, base string) []string {
	links, _ := link.Parse(r) // an error from Parse results in an empty []link being returned
	var ret []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			ret = append(ret, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			ret = append(ret, l.Href)
		}
	}
	return ret
}

func filter(links []string, keepFn func(string) bool) []string {
	var ret []string
	for _, link := range links {
		if keepFn(link) {
			ret = append(ret, link)
		}
	}
	return ret
}

func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}
