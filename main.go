package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

const (
	LIBGEN_URL = "https://libgen.rs/"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Pages     string `json:"pages"`
	Size      string `json:"size"`
	Language  string `json:"language"`
	Category  string `json:"category"`
	Extension string `json:"extension"`
}

type DownloadInfo struct {
	Description string `json:"description"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Author      string `json:"author"`
	Cloudflare  string `json:"cloudflare"`
	IPFSIO      string `json:"ipfsio"`
	Thumbnail   string `json:"thumbnail"`
}

// extractBookInfo extracts book information from a goquery.Selection
func extractBookInfo(s *goquery.Selection) Book {
	return Book{
		Author:    strings.Replace(s.Find("a").First().Text(), "´", "", -1),
		Title:     s.Find("td[width='500']").Find("a").Text(),
		URL:       getLinks(s),
		Pages:     strings.Split(s.Find("td").Eq(5).Text(), "[")[0],
		Size:      s.Find("td").Eq(7).Text(),
		Language:  s.Find("td").Eq(6).Text(),
		Category:  "main",
		Extension: s.Find("td").Eq(8).Text(),
	}
}

// getLinks extracts the download link from a goquery.Selection
func getLinks(s *goquery.Selection) string {
	href, _ := s.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Text() == "[1]"
	}).Attr("href")
	return href
}

// searchEbook performs a search for ebooks and returns a slice of Book structs
func searchEbook(name, page string) ([]Book, error) {
	url := fmt.Sprintf("%ssearch.php?req=%s&page=%s", LIBGEN_URL, strings.ReplaceAll(name, " ", "+"), page)

	resp, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL: %v", errs)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var results []Book
	doc.Find("table.c tr[valign='top']").Each(func(i int, s *goquery.Selection) {
		if i == 0 { // Skip the header row
			return
		}
		results = append(results, extractBookInfo(s))
	})

	return results, nil
}

// getDownload retrieves download information for a given ebook URL
func getDownload(url string) (*DownloadInfo, error) {
	if url == "" {
		return nil, fmt.Errorf("empty URL provided")
	}

	resp, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL: %v", errs)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	info := &DownloadInfo{
		Description: extractDescription(doc),
		Title:       extractTitle(doc),
		URL:         doc.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool { return s.Text() == "GET" }).AttrOr("href", ""),
		Author:      strings.Replace(doc.Find("p").First().Text(), "Author(s): ", "", -1),
		Cloudflare:  doc.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool { return s.Text() == "Cloudflare" }).AttrOr("href", ""),
		IPFSIO:      doc.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool { return s.Text() == "IPFS.io" }).AttrOr("href", ""),
		Thumbnail:   "http://" + strings.Split(url, "/")[2] + doc.Find("img").AttrOr("src", ""),
	}

	return info, nil
}

// extractDescription extracts and cleans up the book description
func extractDescription(doc *goquery.Document) string {
	desc := doc.Find("div").Last().Text()
	desc = strings.TrimSpace(strings.Replace(desc, "Description:", "", -1))
	return strings.Replace(desc, "View a table of contents below:", "", -1)
}

// extractTitle extracts the book title, falling back to textarea content if necessary
func extractTitle(doc *goquery.Document) string {
	title := doc.Find("h1").Text()
	if title == "" {
		textarea := doc.Find("textarea").Text()
		if strings.Contains(textarea, "title =") {
			titlePart := strings.Split(textarea, "title =")[1]
			titlePart = strings.Split(titlePart, "\r")[0]
			title = strings.TrimSpace(strings.Trim(strings.Trim(titlePart, "{"), "},"))
		}
	}
	return title
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	r.GET("/search", handleSearch)
	r.GET("/download", handleDownload)

	r.Run(":8080")
}

// handleSearch handles the /search endpoint
func handleSearch(c *gin.Context) {
	searchTerm := c.Query("ebook")
	page := c.DefaultQuery("page", "1")
	results, err := searchEbook(searchTerm, page)
	if err != nil {
		log.Printf("Error searching ebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search ebook"})
		return
	}
	c.JSON(http.StatusOK, results)
}

// handleDownload handles the /download endpoint
func handleDownload(c *gin.Context) {
	ebookURL := c.Query("ebook")
	result, err := getDownload(ebookURL)
	if err != nil {
		log.Printf("Error downloading ebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download ebook"})
		return
	}
	c.JSON(http.StatusOK, result)
}
