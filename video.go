package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"

	"./index"
)

var videoFilesLocation = "/videos/"

func videoPageHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.TrimPrefix(r.URL.Path, "/video/")
	t, _ := template.ParseFiles("templates/video.html")
	t.Execute(w, videoFilesLocation+fileName)
}

func handleStrippedStaticFiles(prefix string, location string) {
	fileHandler := http.StripPrefix(prefix, http.FileServer(http.Dir(location)))
	loggingHandler := handlers.CombinedLoggingHandler(os.Stdout, fileHandler)
	http.Handle(prefix, loggingHandler)
}

func serve(videoDirectory string, port int) {
	handleStrippedStaticFiles("/static/", "static")
	handleStrippedStaticFiles(videoFilesLocation, videoDirectory)
	http.Handle("/video/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(videoPageHandler)))
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(index.VideoIndexContext{VideoDirectory: videoDirectory}.HandleRequest)))
	listenAddress := fmt.Sprintf(":%d", port)
	http.ListenAndServe(listenAddress, nil)
}

func main() {
	port := flag.Int("port", 8123, "Local port to serve on")
	flag.Parse()
	serve(flag.Arg(0), *port)
}
