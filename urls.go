package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// fetchResult is a struct to hold the result of a fetch operation, including the URL index and content.
type fetchResult struct {
	index   int
	content string
}

// FetchAndCombineContent fetches content from the URLs and combines it in reverse order.
func FetchAndCombineContent(urls []string) (string, error) {
	// Create a new context and channels for this operation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	contentChan := make(chan fetchResult, len(urls))
	errorChan := make(chan error, 1)

	var wg sync.WaitGroup

	// Normalize URLs
	normalizedUrls := make([]string, len(urls))
	for i, url := range urls {
		normalizedUrls[i] = normalizeURL(url)
	}

	// Fetch content from each URL concurrently
	for i, url := range normalizedUrls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			fetchContent(ctx, i, url, contentChan, errorChan)
		}(i, url)
	}

	// Wait for all fetches to complete and then close the content channel
	go func() {
		wg.Wait()
		close(contentChan)
	}()

	// Collect results
	results := make([]string, len(urls))
	for {
		select {
		case result, ok := <-contentChan:
			if !ok {
				// All content has been fetched. Combine the content in reverse order.
				combinedContent := ""
				for i := len(results) - 1; i >= 0; i-- {
					combinedContent += results[i]
				}
				return combinedContent, nil
			}
			results[result.index] = result.content
		case err := <-errorChan:
			// Handle the error and cancel ongoing requests
			return "", err
		}
	}
}

// fetchContent fetches the content from a given URL and sends it to the content channel or error channel.
func fetchContent(ctx context.Context, index int, url string, contentChan chan<- fetchResult, errorChan chan<- error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		errorChan <- fmt.Errorf("failed to create request for %s: %v", url, err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errorChan <- fmt.Errorf("failed to fetch content from %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorChan <- fmt.Errorf("non-success status code %d from %s", resp.StatusCode, url)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorChan <- fmt.Errorf("failed to read content from %s: %v", url, err)
		return
	}
	contentChan <- fetchResult{index: index, content: string(body)}
}

const HTTP string = "http://"
const HTTPS string = "https://"

// normalizeURL ensures that the URL includes the HTTP scheme if it is missing.
func normalizeURL(url string) string {
	if !strings.HasPrefix(url, HTTP) && !strings.HasPrefix(url, HTTPS) {
		return HTTP + url
	}
	return url
}
