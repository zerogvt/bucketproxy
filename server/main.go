package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ListFiles struct {
	FileNames []string `json:"file_names"`
}

type Config struct {
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func listObjects(cfg Config) (error, ListFiles) {
	var lf ListFiles
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region)},
	)
	if err != nil {
		return err, lf
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
		return err, lf
	}
	for _, obj := range result.Contents {
		lf.FileNames = append(lf.FileNames, *obj.Key)
	}
	return nil, lf
}

func main() {
	// Hello world, the web server
	var cfg Config
	configFile, err := os.Open("config.json")
	if err != nil {
		exitErrorf("opening config file", err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&cfg); err != nil {
		exitErrorf("parsing config file", err.Error())
	}
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		err, lf := listObjects(cfg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res, err := json.Marshal(lf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(res))
	}

	http.HandleFunc("/hello", helloHandler)
	log.Println("Listing for requests at http://localhost:8080/hello")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
