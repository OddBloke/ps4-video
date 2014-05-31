package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

var videoFilesLocation = "/videos/"

type videoIndexHandler struct {
	videoDirectory string
}

func (vi videoIndexHandler) handle(w http.ResponseWriter, r *http.Request) {
	var fileNames []string
	fileInfos, _ := ioutil.ReadDir(vi.videoDirectory)
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".mp4") {
			fileNames = append(fileNames, fileInfo.Name())
		}
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, fileNames)
}

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
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(videoIndexHandler{videoDirectory: videoDirectory}.handle)))
	listenAddress := fmt.Sprintf(":%d", port)
	http.ListenAndServe(listenAddress, nil)
}

func main() {
	port := flag.Int("port", 8123, "Local port to serve on")
	flag.Parse()
	serve(flag.Arg(0), *port)
}
