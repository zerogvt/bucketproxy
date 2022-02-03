package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var cfg Config

const FILES_FOLDER = "download"

type ListFiles struct {
	FileNames []string `json:"file_names"`
}

type Config struct {
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
	ServerPort string `json:"server_port"`
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func listObjects(cfg Config) (error, s3.ListObjectsOutput) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region)},
	)
	if err != nil {
		return err, s3.ListObjectsOutput{}
	}
	// Create S3 service client
	svc := s3.New(sess)
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(cfg.BucketName),
		MaxKeys: aws.Int64(2),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err, s3.ListObjectsOutput{}
	}
	return nil, *result
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func ListHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("Listing request. Client: " + ReadUserIP(req))
	err, lf := listObjects(cfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, fmt.Sprintf("%s", lf))
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

func DownloadFromS3(fpath string) error {
	file, err := os.Create(fpath)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", fpath, err)
	}
	defer file.Close()
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region)},
	)

	downloader := s3manager.NewDownloader(sess)
	temp := strings.Split(fpath, "/")
	awsKey := temp[len(temp)-1]
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(cfg.BucketName),
			Key:    aws.String(awsKey),
		})
	if err != nil {
		log.Println("Unable to download item: ", awsKey, err)
		return err
	}
	log.Println("Downloaded", file.Name(), numBytes, "bytes")
	return nil
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
