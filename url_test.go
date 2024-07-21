package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAndCombineContent(t *testing.T) {
	// Set up a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/1":
			w.Write([]byte("some content 1\n"))
		case "/2":
			w.Write([]byte("some content 2\n"))
		case "/3":
			w.Write([]byte("some content 3\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Test cases
	tests := []struct {
		name     string
		urls     []string
		expected string
		hasError bool
	}{
		{
			name:     "Valid URLs",
			urls:     []string{server.URL + "/1", server.URL + "/2", server.URL + "/3"},
			expected: "some content 3\nsome content 2\nsome content 1\n",
			hasError: false,
		},
		{
			name:     "Mixed valid and invalid URLs",
			urls:     []string{server.URL + "/1", "invalid-url", server.URL + "/3"},
			expected: "",
			hasError: true,
		},
		{
			name:     "Reverse order",
			urls:     []string{server.URL + "/3", server.URL + "/2", server.URL + "/1"},
			expected: "some content 1\nsome content 2\nsome content 3\n",
			hasError: false,
		},
		{
			name:     "Empty URL list",
			urls:     []string{},
			expected: "",
			hasError: false,
		},
	}

	fetcher := NewURLFetcher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fetcher.FetchAndCombineContent(tt.urls)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}
