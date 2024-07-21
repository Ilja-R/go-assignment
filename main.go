package main

import (
	"log"
	"net/http"
	"os"
)

// URLS Some hardcoded URLs to fetch content from as an example
var URLS = []string{
	"raw.githubusercontent.com/GoogleContainerTools/distroless/main/java/README.md",
	"raw.githubusercontent.com/golang/go/master/README.md",
}

func getUrlContentsHandler(w http.ResponseWriter, _ *http.Request) {
	content, err := FetchAndCombineContent(URLS)
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
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	log.Println("Server started at http://localhost:" + httpPort)
	log.Println("Use http://localhost:" + httpPort + "/getUrlContents to fetch and combine content from URLs")
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
