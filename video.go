package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/handlers"
)

var videoFilesLocation = "/videos/"

type videoFile struct {
	FileName          string
	ThumbnailFileName string
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

func (vi videoIndexHandler) generateThumbnail(videoFileName string) string {
	videoHash := sha1.Sum([]byte(videoFileName))
	expectedFilename := fmt.Sprintf("%s/%x.png", vi.videoDirectory, videoHash)
	videoPath := vi.videoDirectory + "/" + videoFileName
	thumbnailGenerationCommand := exec.Command(
		"totem-video-thumbnailer", "-s", "640", videoPath, expectedFilename)
	err := thumbnailGenerationCommand.Run()
	if err == nil {
		return expectedFilename
	}
	return ""
}

func (vi videoIndexHandler) getVideoThumbnailFileName(videoFileName string) string {
	videoHash := sha1.Sum([]byte(videoFileName))
	expectedFilename := fmt.Sprintf("%s/%x.png", vi.videoDirectory, videoHash)
	_, err := os.Open(expectedFilename)
	if err == nil {
		return fmt.Sprintf("%x.png", videoHash)
	}
	return vi.generateThumbnail(videoFileName)
}

func (vi videoIndexHandler) handle(w http.ResponseWriter, r *http.Request) {
	var videoFiles []videoFile
	fileInfos, _ := ioutil.ReadDir(vi.videoDirectory)
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".mp4") {
			thumbnail := vi.getVideoThumbnailFileName(fileInfo.Name())
			file := videoFile{FileName: fileInfo.Name(), ThumbnailFileName: thumbnail}
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
