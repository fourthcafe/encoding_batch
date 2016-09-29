package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var pathSeparator = string(filepath.Separator)
var availableExtension = []string{".mp4", ".mkv", ".avi"}

type EncodingInfo struct {
	size, bitrate string
}

var encodingInfoMap = map[string]EncodingInfo{
	"1080p": {size: "1920:-2", bitrate: "4000k"},
	"720p":  {size: "1280:-2", bitrate: "2500k"},
	"360p":  {size: "640:-2", bitrate: "700k"},
	"180p":  {size: "320:-2", bitrate: "400k"},
}

type PathInfo struct {
	root, sub string
}

type VideoInfo struct {
	title, quality string
}

// 인코딩한 파일을 저장할 현재 경로의 하위 폴더명
var ENCODING_DEST = "encoding"
var path = PathInfo{}

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln("Err Occure: Can not read Dir - ", os.Args[0])
	}

	path = PathInfo{
		root: dir,
		sub:  ENCODING_DEST,
	}

	encodingDir := filepath.Join(path.root, path.sub)
	if _, err := os.Stat(encodingDir); os.IsNotExist(err) {
		if err := os.Mkdir(encodingDir, 0754); err != nil {
			log.Println("fail to make directory:", encodingDir)
			log.Panicln(err.Error())
		} else {
			log.Println("make directory:", encodingDir)
		}
	}

	LOG_PATH := filepath.Join(encodingDir, "log.txt")
	f, err := os.OpenFile(LOG_PATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModeAppend)
	if err == nil {
		log.SetOutput(f)
	}
}

func main() {
	files, err := ioutil.ReadDir(path.root)
	if err != nil {
		panic(err.Error())
	}

	for _, file := range files {
		if file.IsDir() {
			// log.Println("d: ", file.Name())

		} else {
			if isContain(availableExtension, filepath.Ext(file.Name())) {
				encodingOrigin(file.Name())
			}
		}
	}
}

func isContain(strArr []string, str string) bool {
	for _, value := range strArr {
		if str == value {
			return true
		}
	}

	return false
}

// 최초 원본 파일 인코딩
func encodingOrigin(originFileName string) {
	video := extractVideoInfo(originFileName)

	encodFileName := video.title + "_" + video.quality + ".mp4"
	ffmpegEncoding(originFileName, encodFileName, video.quality, true, true)

	encodingRowQuality(encodFileName)
}

func extractVideoInfo(fileName string) VideoInfo {
	extension := filepath.Ext(fileName)
	nonExtName := fileName[0:strings.LastIndex(fileName, extension)]
	fileSeparateIdx := strings.LastIndex(nonExtName, "_")

	video := VideoInfo{}
	video.title = nonExtName[0:fileSeparateIdx]
	video.quality = nonExtName[fileSeparateIdx+1:]

	return video
}

// 하위 해상도를 인코딩한다
func encodingRowQuality(originFileName string) {
	video := extractVideoInfo(originFileName)

	isFirst := false
	hasWatermark := false

	switch video.quality {
	case "1080p":
		ffmpegEncoding(originFileName, video.title+"_720p.mp4", "720p", isFirst, hasWatermark)
		fallthrough
	case "720p":
		ffmpegEncoding(originFileName, video.title+"_360p.mp4", "360p", isFirst, hasWatermark)
		fallthrough
	case "360p":
		ffmpegEncoding(originFileName, video.title+"_180p.mp4", "180p", isFirst, hasWatermark)
	default:
		log.Println("no more encoding file: ", originFileName)
	}
}

// 실제 인코딩
func ffmpegEncoding(oldFile, newFile, quality string, isFirst, hasWatermark bool) {
	defer func() {
		recover()
	}()

	oldFilePath := ""
	if isFirst {
		oldFilePath = filepath.Join(path.root, oldFile)

	} else {
		oldFilePath = filepath.Join(path.root, path.sub, oldFile)
	}
	newFilePath := filepath.Join(path.root, path.sub, newFile)

	encodingInfo := encodingInfoMap[quality]
	start := time.Now()

	cmd := &exec.Cmd{}
	if hasWatermark {
		cmd = exec.Command("C:/Temp/FFmpeg/bin/ffmpeg",
			"-i", oldFilePath,
			"-vf",
			"scale="+encodingInfo.size+
				", drawtext=fontsize=30:fontfile=C:/Temp/FFmpeg/impact.ttf:text=Video ini hanya digunakan untuk tes.:x=(w-text_w)/2:y=20:fontcolor=white:enable=lt(mod(t\\,60)\\,10):",
			"-maxrate", encodingInfo.bitrate,
			"-b:a", "128k",
			"-ac", "2",
			newFilePath,
		)
	} else {
		cmd = exec.Command("C:/Temp/FFmpeg/bin/ffmpeg",
			"-i", oldFilePath,
			"-vf",
			"scale="+encodingInfo.size,
			"-maxrate", encodingInfo.bitrate,
			"-b:a", "128k",
			"-ac", "2",
			newFilePath,
		)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Panicln(err.Error())
	}

	log.Printf("=============== Now encoding... ===============")
	log.Printf("encoding: %s -> %s\n", oldFilePath, newFilePath)

	err = cmd.Wait()
	if err != nil {
		log.Println("err occure in encoding")
		log.Println(oldFilePath, "encoding to", newFilePath)
		log.Panicln(err.Error())
	}

	log.Printf("complate: %s %s\n", newFilePath, time.Since(start).String())
}
