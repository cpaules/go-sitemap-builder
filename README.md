Outputs a sitemap in the XML format specified by the [standard sitemap protocol](https://www.sitemaps.org/index.html).

Running ```
go run main.go
``` will print the sitemap to the console. Accepts a `url` flag and a `depth` flag, which default to `https://www.google.com` and to `2`

Creates a sitemap by visiting the base url and performing a breadth first search on each link in the page until the sitemap 'tree' is exhausted or the maximum depth set by the flag or the default value of 2 is reached.