package router

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

var workDir string

func AddRoutes(
	mux *http.ServeMux,
) {
	workDir, _ = os.Getwd()
	mux.HandleFunc("/", handleHTML)
	// mux.HandleFunc("/api/v1/stats", handleServerStats)
}

func loadHTML(filePath string) []byte {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}
	return data
}

func handleHTML(
	w http.ResponseWriter,
	r *http.Request,
) {
	var data []byte
	switch r.URL.Path {
	// case "<URL/ROUTE>":
	// 	data = loadHTML("../app/<FilePath>")
	case "/":
		data = loadHTML(workDir + "/app/index.html")
	default:
		data = loadHTML(workDir + "/app/http_error/404.html")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func handleServerStats(
	w http.ResponseWriter,
	r *http.Request,
) {
	var data []byte

	cpuCores := "Number of CPU's: " + strconv.Itoa(runtime.NumCPU()) + "\n"
	goRoutines := "Number of Go Routine's: " + strconv.Itoa(runtime.NumGoroutine()) + "\n"
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	mallocs := "Memory Allocated: " + strconv.Itoa(int(mem.Mallocs)) + "B\n"
	heapAlloc := "Heap Allocated: " + strconv.Itoa(int(mem.HeapAlloc)) + "B\n"
	heapInUse := "Heap In Use: " + strconv.Itoa(int(mem.HeapInuse)) + "B\n"

	data = []byte(cpuCores + goRoutines + mallocs + heapAlloc + heapInUse)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}
