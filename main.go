package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/gocolly/colly"
)

func main() {

	var URL string = ""

	flag.StringVar(&URL, "u", "", "-u")

	flag.Parse()

	ScrapeLinks(URL)

}

func ScrapeLinks(Url string) {

	u := GetUrlInfo(Url)

	if !u.IsAbsolute {
		panic("Invalid Url")
	}

	config, err := getSiteConfig(u.Domain)

	if err != nil {
		panic("Site not found")
	}

	result := ExtractedResult{
		Info:  u,
		Links: []ExtractedLink{},
	}

	c := colly.NewCollector(
		colly.AllowedDomains(u.Domain),
	)

	c.OnRequest(func(r *colly.Request) {

		fmt.Println("Visiting ", u.URL)
	})

	c.OnHTML(config.Selector, func(h *colly.HTMLElement) {

		href := h.Attr("href")
		text := h.Text
		link := GetAbsoluteUrl(href, u)
		extractedLink := ExtractedLink{
			Url:  link,
			Text: text,
		}
		result.Links = append(result.Links, extractedLink)

	})

	c.OnError(func(r *colly.Response, err error) {

		fmt.Println(err)

	})

	c.OnScraped(func(r *colly.Response) {

		fmt.Println("Scrape completed")

		result.SaveResult("list")

	})

	c.Visit(u.URL)

}

type URLInfo struct {
	Domain     string
	BaseURL    string
	URL        string
	IsAbsolute bool
}

func GetUrlInfo(Url string) URLInfo {

	u, err := url.Parse(Url)

	if err != nil {
		panic(err)
	}

	return URLInfo{
		Domain:     u.Host,
		BaseURL:    fmt.Sprintf("%s://%s", u.Scheme, u.Host),
		URL:        u.String(),
		IsAbsolute: u.IsAbs(),
	}
}

func GetAbsoluteUrl(urlPath string, u URLInfo) string {

	info := GetUrlInfo(urlPath)

	if !info.IsAbsolute {

		return fmt.Sprintf("%s%s", u.BaseURL, urlPath)
	}

	return info.URL
}

type ExtractedLink struct {
	Url  string
	Text string
}

type ExtractedResult struct {
	Info  URLInfo
	Links []ExtractedLink
}

func (result *ExtractedResult) SaveResult(filename string) {

	var content string = fmt.Sprintf("# URL  : %s \n\n", result.Info.URL)

	for _, value := range result.Links {

		content = content + fmt.Sprintf("\n # %s\n", value.Text)
		content = content + fmt.Sprintf("%s\n", value.Url)

	}

	file, error := os.Create(filename)

	if error != nil {
		fmt.Printf("Creating %s error ", filename)
		fmt.Println(error)
		return
	}

	defer file.Close()

	_, writeError := file.WriteString(content)

	if writeError != nil {
		fmt.Println("File writing errr ", writeError)
	}

}

func getSiteConfig(domain string) (ScrapeConfig, error) {

	config, isOK := Sites[domain]

	if !isOK {
		return config, errors.New("site config not found")
	}

	return config, nil

}

var Sites map[string]ScrapeConfig = map[string]ScrapeConfig{"egghead.io": {Selector: "ul li a.text-lg.font-semibold"}}

type ScrapeConfig struct {
	Selector string
}
