package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

var VideoFilesLocation = "/videos/"

func video_page_handler(w http.ResponseWriter, r *http.Request) {
	file_name := strings.TrimPrefix(r.URL.Path, "/video/")
	t, _ := template.ParseFiles("video.html")
	t.Execute(w, VideoFilesLocation+file_name)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	var file_names []string
	video_directory := os.Args[1]
	file_infos, _ := ioutil.ReadDir(video_directory)
	for _, file_info := range file_infos {
		if strings.HasSuffix(file_info.Name(), ".mp4") {
			file_names = append(file_names, file_info.Name())
		}
	}
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, file_names)
}

func handle_stripped_static_files(prefix string, location string) {
	file_handler := http.StripPrefix(prefix, http.FileServer(http.Dir(location)))
	logging_handler := handlers.CombinedLoggingHandler(os.Stdout, file_handler)
	http.Handle(prefix, logging_handler)
}

func main() {
	handle_stripped_static_files("/static/", "static")
	handle_stripped_static_files(VideoFilesLocation, os.Args[1])
	http.Handle("/video/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(video_page_handler)))
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(index_handler)))
	http.ListenAndServe(":8123", nil)
}
