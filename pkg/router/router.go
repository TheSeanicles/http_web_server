package router

import (
	"log"
	"net/http"
	"os"
)

func AddRoutes(
	mux *http.ServeMux,
) {
	mux.Handle("/api/v1", http.NotFoundHandler())
	mux.HandleFunc("/", handleHTML)
}

func loadHTML(filePath string) []byte {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}
	return data
}

func handleHTML(
	// logger
	w http.ResponseWriter,
	r *http.Request,
) {
	var data []byte
	switch r.URL.Path {
	case "/":
		data = loadHTML("../app/index.html")
	case "/about":
		data = loadHTML("../app/about/index.html")
	default:
		data = loadHTML("../app/http_error/404.html")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}
