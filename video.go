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

func main() {
	static_handler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	http.Handle("/static/", handlers.CombinedLoggingHandler(os.Stdout, static_handler))
	http.Handle("/video/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(video_page_handler)))
	http.Handle(VideoFilesLocation, handlers.CombinedLoggingHandler(os.Stdout, http.StripPrefix(VideoFilesLocation, http.FileServer(http.Dir(os.Args[1])))))
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(index_handler)))
	http.ListenAndServe(":8123", nil)
}
