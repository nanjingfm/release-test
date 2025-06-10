package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/time/rate"
)

// HTTPClient wraps http.Client with rate limiting functionality
type HTTPClient struct {
	client  *http.Client
	limiter *rate.Limiter
}

// NewHTTPClient creates a new HTTP client with rate limiting
// limit: requests per second, burst: maximum burst size
func NewHTTPClient(limit rate.Limit, burst int) *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		limiter: rate.NewLimiter(limit, burst),
	}
}

// Get performs a rate-limited HTTP GET request
func (c *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	// Wait for rate limiter permission
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Perform the request
	return c.client.Do(req)
}

// fetchAndParseHTML fetches HTML content from the given URL and extracts title
func fetchAndParseHTML(url string) (string, error) {
	// Create HTTP client with rate limiting (1 request per second, burst of 3)
	httpClient := NewHTTPClient(rate.Every(1*time.Second), 3)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("等待限流器许可...")

	// Fetch the webpage content with rate limiting
	resp, err := httpClient.Get(ctx, url)
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
	// Example URLs to fetch and parse
	urls := []string{
		"https://golang.org",
		"https://pkg.go.dev",
	}

	for _, url := range urls {
		fmt.Printf("正在获取并解析网页: %s\n", url)

		title, err := fetchAndParseHTML(url)
		if err != nil {
			log.Printf("Error fetching %s: %v", url, err)
			continue
		}

		if title != "" {
			fmt.Printf("网页标题: %s\n", title)
		} else {
			fmt.Println("未找到网页标题")
		}
		fmt.Println("---")
	}
}
