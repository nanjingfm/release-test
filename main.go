package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// fetchAndParseHTML fetches HTML content from the given URL and extracts title
func fetchAndParseHTML(url string) (string, error) {
	// Fetch the webpage content
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Parse the HTML content
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract the title from the parsed HTML
	title := extractTitle(doc)
	return title, nil
}

// extractTitle traverses the HTML tree to find the title element
func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		return getTextContent(n.FirstChild)
	}

	// Recursively search through child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}
	return ""
}

// getTextContent extracts text content from a node
func getTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}
	return ""
}

func main() {
	// Example URL to fetch and parse
	url := "https://golang.org"

	fmt.Printf("正在获取并解析网页: %s\n", url)

	title, err := fetchAndParseHTML(url)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if title != "" {
		fmt.Printf("网页标题: %s\n", title)
	} else {
		fmt.Println("未找到网页标题")
	}
}
