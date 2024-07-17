package main

import (
	"log"
	"net/http"
)

var URLS = []string{
	"raw.githubusercontent.com/GoogleContainerTools/distroless/main/java/README.md",
	"raw.githubusercontent.com/golang/go/master/README.md",
}

func getUrlContentsHandler(w http.ResponseWriter, _ *http.Request) {
	fetcher := NewURLFetcher()
	content, err := fetcher.FetchAndCombineContent(URLS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func main() {
	http.HandleFunc("/getUrlContents", getUrlContentsHandler)
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
