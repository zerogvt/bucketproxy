package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type ListFiles struct {
	FileNames []string `json:"file_names"`
}

func DownloadFromS3(fpath string) error {
	file, err := os.Create(fpath)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", fpath, err)
	}
	defer file.Close()
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

func listObjects(cfg Config) (s3.ListObjectsOutput, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region)},
	)
	if err != nil {
		return s3.ListObjectsOutput{}, err
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
		return s3.ListObjectsOutput{}, err
	}
	return *result, nil
}
