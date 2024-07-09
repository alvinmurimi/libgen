package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
)

const (
	LIBGEN_URL = "https://libgen.rs/"
)

type Book struct {
	Authors   []string `json:"authors"`
	Title     string   `json:"title"`
	Series    string   `json:"series,omitempty"`
	Edition   string   `json:"edition,omitempty"`
	ISBNs     []string `json:"isbns,omitempty"`
	URL       string   `json:"url"`
	Publisher string   `json:"publisher"`
	Year      string   `json:"year"`
	Pages     string   `json:"pages"`
	Size      string   `json:"size"`
	Language  string   `json:"language"`
	Category  string   `json:"category"`
	Extension string   `json:"extension"`
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

func extractBookInfo(s *goquery.Selection) Book {
	book := Book{
		Category: "main",
		ISBNs:    []string{},
	}

	// Extract authors
	s.Find("td").Eq(1).Find("a").Each(func(i int, s *goquery.Selection) {
		book.Authors = append(book.Authors, strings.TrimSpace(s.Text()))
	})

	// Extract title, series, edition, and ISBNs
	titleCell := s.Find("td[width='500']")
	titleLink := titleCell.Find("a")
	book.Title = titleLink.Text()

	// Extract series if present
	seriesFont := titleCell.Find("font[face='Times'][color='green']").First()
	if seriesFont.Length() > 0 && !strings.Contains(seriesFont.Text(), "ISBN") {
		book.Series = seriesFont.Text()
	}

	// Extract edition if present
	editionFont := titleCell.Find("font[face='Times'][color='green']").Eq(1)
	if editionFont.Length() > 0 && !strings.Contains(editionFont.Text(), "ISBN") {
		book.Edition = editionFont.Text()
	}

	// Extract ISBNs
	isbnText := titleCell.Find("font[face='Times'][color='green']").Last().Text()
	book.ISBNs = extractISBNs(isbnText)

	// Extract other information
	book.URL = getLinks(s)
	book.Publisher = strings.TrimSpace(s.Find("td").Eq(3).Text())
	book.Year = strings.TrimSpace(s.Find("td").Eq(4).Text())
	book.Pages = strings.Split(s.Find("td").Eq(5).Text(), "[")[0]
	book.Language = s.Find("td").Eq(6).Text()
	book.Size = s.Find("td").Eq(7).Text()
	book.Extension = s.Find("td").Eq(8).Text()

	return book
}

func getLinks(s *goquery.Selection) string {
	href, _ := s.Find("td a").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Text() == "[1]"
	}).Attr("href")
	return href
}

func extractISBNs(text string) []string {
	isbns := []string{}
	for _, item := range strings.Split(text, ", ") {
		if isISBN(item) {
			isbns = append(isbns, item)
		}
	}
	return isbns
}

func isISBN(s string) bool {
	// Remove any hyphens or spaces
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")

	// Check if it's a valid ISBN-10 or ISBN-13
	return isISBN10(s) || isISBN13(s)
}

func isISBN10(s string) bool {
	if len(s) != 10 {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(s[i]))
		if err != nil {
			return false
		}
		sum += digit * (10 - i)
	}
	last := s[9]
	if last == 'X' {
		sum += 10
	} else {
		digit, err := strconv.Atoi(string(last))
		if err != nil {
			return false
		}
		sum += digit
	}
	return sum%11 == 0
}

func isISBN13(s string) bool {
	if len(s) != 13 {
		return false
	}
	sum := 0
	for i := 0; i < 12; i++ {
		digit, err := strconv.Atoi(string(s[i]))
		if err != nil {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	check, err := strconv.Atoi(string(s[12]))
	if err != nil {
		return false
	}
	return (10-(sum%10))%10 == check
}

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

func extractDescription(doc *goquery.Document) string {
	desc := doc.Find("div").Last().Text()
	desc = strings.TrimSpace(strings.Replace(desc, "Description:", "", -1))
	return strings.Replace(desc, "View a table of contents below:", "", -1)
}

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
