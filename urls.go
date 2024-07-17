package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const HTTP string = "http://"
const HTTPS string = "https://"

type URLFetcher interface {
	FetchAndCombineContent(urls []string) (string, error)
}

type urlFetcher struct {
	ContentChan chan fetchResult
	ErrorChan   chan error
	ctx         context.Context
	cancel      context.CancelFunc
}

// This is required to persist an order for async code
type fetchResult struct {
	index   int
	content string
}

func NewURLFetcher() URLFetcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &urlFetcher{
		ContentChan: make(chan fetchResult),
		ErrorChan:   make(chan error, 1),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (uf *urlFetcher) FetchAndCombineContent(urls []string) (string, error) {
	var wg sync.WaitGroup
	uf.ContentChan = make(chan fetchResult, len(urls))

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

	results := make([]string, len(urls))
	for {
		select {
		case result, ok := <-uf.ContentChan:
			if !ok {
				combinedContent := ""
				for i := len(results) - 1; i >= 0; i-- {
					combinedContent += results[i]
				}
				return combinedContent, nil
			}
			results[result.index] = result.content
		case err := <-uf.ErrorChan:
			uf.cancel() // Cancel all ongoing requests
			return "", err
		}
	}
}

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

// adds http:// or https:// if url does not start with it
func normalizeURL(url string) string {
	if !strings.HasPrefix(url, HTTP) && !strings.HasPrefix(url, HTTPS) {
		return HTTP + url
	}
	return url
}
