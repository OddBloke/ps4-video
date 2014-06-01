package index

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type VideoIndexContext struct {
	VideoDirectory string
	VideoURLPrefix string
}

type videoFilename string

func (v videoFilename) Hash() string {
	videoHash := sha1.Sum([]byte(v))
	return fmt.Sprintf("%x", videoHash)
}

func (v videoFilename) FullPath(context VideoIndexContext) string {
	return fmt.Sprintf("%s/%s", context.VideoDirectory, v)
}

type videoFile struct {
	FileName  videoFilename
	Thumbnail indexThumbnail
}

func NewVideoFile(context VideoIndexContext, fileName string) videoFile {
	videoName := videoFilename(fileName)
	thumbnail := CreateThumbnail(context, videoName)
	return videoFile{videoName, thumbnail}
}

type indexThumbnail struct {
	context            VideoIndexContext
	videoFile          videoFilename
	fileSystemLocation string
}

func CreateThumbnail(context VideoIndexContext, video videoFilename) indexThumbnail {
	fileSystemLocation := fmt.Sprintf("%s/%s.png", context.VideoDirectory, video.Hash())
	return indexThumbnail{context, video, fileSystemLocation}
}

func (t indexThumbnail) GetURL() string {
	expectedURL := t.context.VideoURLPrefix + t.videoFile.Hash() + ".png"
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

func (context VideoIndexContext) HandleRequest(w http.ResponseWriter, r *http.Request) {
	var videoFiles []videoFile
	fileInfos, _ := ioutil.ReadDir(context.VideoDirectory)
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".mp4") {
			file := NewVideoFile(context, fileInfo.Name())
			videoFiles = append(videoFiles, file)
		}
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, makeVideoRows(videoFiles, 3))
}
