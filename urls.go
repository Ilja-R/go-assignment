package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// URLFetcher defines an interface for fetching and combining URL contents
type URLFetcher interface {
	FetchAndCombineContent(urls []string) (string, error)
}

type urlFetcher struct {
	ContentChan chan fetchResult
	ErrorChan   chan error
	ctx         context.Context
	cancel      context.CancelFunc
}

// fetchResult is a struct to hold the result of a fetch operation. Important goal is to keep the index of the URL.
type fetchResult struct {
	index   int
	content string
}

// NewURLFetcher creates a new urlFetcher instance
func NewURLFetcher() URLFetcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &urlFetcher{
		ContentChan: make(chan fetchResult),
		ErrorChan:   make(chan error, 1),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// FetchAndCombineContent fetches content from the URLs and combines it in reverse order.
func (uf *urlFetcher) FetchAndCombineContent(urls []string) (string, error) {
	// Reset channels and context for reuse of the fetcher
	uf.ContentChan = make(chan fetchResult, len(urls))
	uf.ErrorChan = make(chan error, 1)
	uf.ctx, uf.cancel = context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Normalize URLs. Ensure they start with http:// or https://
	normalizedUrls := make([]string, len(urls))
	for i, url := range urls {
		normalizedUrls[i] = normalizeURL(url)
	}

	for i, url := range normalizedUrls {
		wg.Add(1)
		go uf.fetchContent(i, url, &wg)
	}

	go func() {
		wg.Wait()
		close(uf.ContentChan)
	}()

	// Collect the content from the channel or handle the error
	results := make([]string, len(urls))
	for {
		select {
		case result, ok := <-uf.ContentChan:
			if !ok {
				// All content has been fetched. Combine the content in reverse order.
				combinedContent := ""
				for i := len(results) - 1; i >= 0; i-- {
					combinedContent += results[i]
				}
				return combinedContent, nil
			}
			// Here is the place where put the content in the right order.
			results[result.index] = result.content
		case err := <-uf.ErrorChan:
			uf.cancel() // Cancel all ongoing requests.
			return "", err
		}
	}
}

// fetchContent fetches the content from a given URL and sends it to the content channel or error channel
func (uf *urlFetcher) fetchContent(index int, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequestWithContext(uf.ctx, "GET", url, nil)
	if err != nil {
		uf.ErrorChan <- fmt.Errorf("failed to create request for %s: %v", url, err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		uf.ErrorChan <- fmt.Errorf("failed to fetch content from %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		uf.ErrorChan <- fmt.Errorf("non-success status code %d from %s", resp.StatusCode, url)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uf.ErrorChan <- fmt.Errorf("failed to read content from %s: %v", url, err)
		return
	}
	uf.ContentChan <- fetchResult{index: index, content: string(body)}
}

const HTTP string = "http://"
const HTTPS string = "https://"

// normalizeURL ensures that the URL includes the http scheme if it is missing.
func normalizeURL(url string) string {
	// Maybe needed a more robust implementation. This is just according to provided example.
	if !strings.HasPrefix(url, HTTP) && !strings.HasPrefix(url, HTTPS) {
		return HTTP + url
	}
	return url
}
