package main

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type shortenerURLStorage map[string]string

func (storage shortenerURLStorage) AddURL(url string) string {
	var urlID string
	addition := ""
	for {
		urlID = shortener(url + addition)

		if v, ok := storage[urlID]; !ok {
			storage[urlID] = url
			break
		} else if v == url {
			break
		} else {
			addition += "1"
		}
	}
	return urlID
}

var storage shortenerURLStorage = shortenerURLStorage{}

func shortener(url string) string {
	hash := sha256.Sum256([]byte(url))
	result := []byte(base32.StdEncoding.EncodeToString(hash[:]))[:8]

	return string(result)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/{id}", redirectHandler)
	fmt.Println("Сервер запущен на :8080")
	return http.ListenAndServe(`:8080`, mux)
}

func rootHandler(rw http.ResponseWriter, req *http.Request) {

	fmt.Printf("Content-type: %s\n", req.Header.Get("content-type"))
	fmt.Printf("url: %s\n", req.URL.Path)
	fmt.Printf("method: %s\n", req.Method)

	contentType := req.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(rw, "Only POST is allowed", http.StatusBadRequest)
		return
	}
	if req.URL.Path != "/" {
		http.Error(rw, "URL is not allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}

	fmt.Printf("Body: %s\n", body)
	url := string(body)
	urlID := storage.AddURL(url)

	rw.Header().Set("content-type", "text/plain")
	rw.WriteHeader(http.StatusCreated)

	rw.Write([]byte("http://localhost:8080/" + urlID))

}

func redirectHandler(rw http.ResponseWriter, req *http.Request) {

	fmt.Printf("Content-type: %s\n", req.Header.Get("content-type"))
	fmt.Printf("url: %s\n", req.URL.Path)
	fmt.Printf("method: %s\n", req.Method)

	if req.Method != http.MethodGet {

		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	id := req.PathValue("id")

	url, ok := storage[id]
	if !ok {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}

	rw.Header().Set("Location", url)
	rw.WriteHeader(http.StatusTemporaryRedirect)
}
