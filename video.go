package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)


type VideoPage struct {
	Host string
	FileName string
}


func video(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")[0]
	file_name := r.URL.Path[1:]
	t, _ := template.ParseFiles("video.html")
	t.Execute(w, VideoPage{Host: host, FileName: file_name})
}


func index(w http.ResponseWriter, r *http.Request) {
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

type MyHandler struct {}

func (h MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if (r.URL.Path == "/") {
		index(w, r)
	} else {
		video(w, r)
	}
}

func main() {
	static_handler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	http.Handle("/static/", handlers.CombinedLoggingHandler(os.Stdout, static_handler))
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, MyHandler{}))
	http.ListenAndServe(":8123", nil)
}
