package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var cfg Config

const FILES_FOLDER = "download"

type Config struct {
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
	ServerPort string `json:"server_port"`
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func ListHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("Listing request. Client: " + ReadUserIP(req))
	lf, err := listObjects(cfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, lf.GoString())
}

func GetFileHandler(w http.ResponseWriter, req *http.Request) {
	target := req.URL.Path[1:]
	log.Println("Serving: " + target + " Client: " + ReadUserIP(req))
	if _, err := os.Stat(target); err != nil {
		// file not locally - download from S3
		erraws := DownloadFromS3(target)
		if erraws != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	http.ServeFile(w, req, target)
}

func main() {
	err := os.MkdirAll(FILES_FOLDER, 0700)
	if err != nil {
		log.Fatal(err)
	}
	configFile, err := os.Open("config.json")
	if err != nil {
		exitErrorf("opening config file", err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&cfg); err != nil {
		exitErrorf("parsing config file", err.Error())
	}

	http.HandleFunc("/list", ListHandler)
	http.HandleFunc("/"+FILES_FOLDER+"/", GetFileHandler)

	log.Println("Listening at http://localhost:" + cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}
