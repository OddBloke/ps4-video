package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

var videoFilesLocation = "/videos/"

func videoPageHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.TrimPrefix(r.URL.Path, "/video/")
	t, _ := template.ParseFiles("video.html")
	t.Execute(w, videoFilesLocation+fileName)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var fileNames []string
	videoDirectory := os.Args[1]
	fileInfos, _ := ioutil.ReadDir(videoDirectory)
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".mp4") {
			fileNames = append(fileNames, fileInfo.Name())
		}
	}
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, fileNames)
}

func handleStrippedStaticFiles(prefix string, location string) {
	fileHandler := http.StripPrefix(prefix, http.FileServer(http.Dir(location)))
	loggingHandler := handlers.CombinedLoggingHandler(os.Stdout, fileHandler)
	http.Handle(prefix, loggingHandler)
}

func main() {
	handleStrippedStaticFiles("/static/", "static")
	handleStrippedStaticFiles(videoFilesLocation, os.Args[1])
	http.Handle("/video/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(videoPageHandler)))
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(indexHandler)))
	http.ListenAndServe(":8123", nil)
}
