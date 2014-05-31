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

type videoFile struct {
	FileName string
}

type videoIndexHandler struct {
	videoDirectory string
}

func makeVideoRows(videoFiles []videoFile, rowLength int) [][]videoFile {
	var rows [][]videoFile
	var new_row []videoFile
	for _, file := range videoFiles {
		if len(new_row) == rowLength {
			rows = append(rows, new_row)
			new_row = []videoFile{file}
		} else {
			new_row = append(new_row, file)
		}
	}
	if len(new_row) > 0 {
		rows = append(rows, new_row)
	}
	return rows
}

func (vi videoIndexHandler) handle(w http.ResponseWriter, r *http.Request) {
	var videoFiles []videoFile
	fileInfos, _ := ioutil.ReadDir(vi.videoDirectory)
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".mp4") {
			file := videoFile{fileInfo.Name()}
			videoFiles = append(videoFiles, file)
		}
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, makeVideoRows(videoFiles, 3))
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
