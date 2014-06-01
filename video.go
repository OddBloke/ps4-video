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

type videoFilename string

func (v videoFilename) Hash() string {
	videoHash := sha1.Sum([]byte(v))
	return fmt.Sprintf("%x", videoHash)
}

func (v videoFilename) FullPath(context videoIndexContext) string {
	return fmt.Sprintf("%s/%s", context.videoDirectory, v)
}

type videoFile struct {
	FileName  videoFilename
	Thumbnail indexThumbnail
}

func NewVideoFile(context videoIndexContext, fileName string) videoFile {
	videoName := videoFilename(fileName)
	thumbnail := CreateThumbnail(context, videoName)
	return videoFile{videoName, thumbnail}
}

type videoIndexContext struct {
	videoDirectory string
}

type indexThumbnail struct {
	context            videoIndexContext
	videoFile          videoFilename
	fileSystemLocation string
}

func CreateThumbnail(context videoIndexContext, video videoFilename) indexThumbnail {
	fileSystemLocation := fmt.Sprintf("%s/%s.png", context.videoDirectory, video.Hash())
	return indexThumbnail{context, video, fileSystemLocation}
}

func (t indexThumbnail) GetURL() string {
	expectedURL := videoFilesLocation + t.videoFile.Hash() + ".png"
	_, err := os.Open(t.fileSystemLocation)
	if err == nil {
		return expectedURL
	}
	thumbnailGenerationCommand := exec.Command(
		"totem-video-thumbnailer", "-s", "640", t.videoFile.FullPath(t.context), t.fileSystemLocation)
	err = thumbnailGenerationCommand.Run()
	if err == nil {
		return expectedURL
	}
	return ""
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

func (vi videoIndexContext) handleRequest(w http.ResponseWriter, r *http.Request) {
	var videoFiles []videoFile
	fileInfos, _ := ioutil.ReadDir(vi.videoDirectory)
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".mp4") {
			file := NewVideoFile(vi, fileInfo.Name())
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
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(videoIndexContext{videoDirectory: videoDirectory}.handleRequest)))
	listenAddress := fmt.Sprintf(":%d", port)
	http.ListenAndServe(listenAddress, nil)
}

func main() {
	port := flag.Int("port", 8123, "Local port to serve on")
	flag.Parse()
	serve(flag.Arg(0), *port)
}
