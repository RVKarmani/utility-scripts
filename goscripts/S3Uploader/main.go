package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fatih/color"
	"github.com/ncruces/zenity"
)

const defaultPath = ``

const S3_BUCKET = "[S3_BUCKET]"
const AWS_REGION = "[AWS_REGION]"

func main() {
	uploadFilePath, _ := zenity.SelectFile(
		zenity.Filename(defaultPath), zenity.FileFilters{{"CSV files", []string{"*.csv"}}, {"All files", []string{"*"}}},
	)

	uploadFileName := filepath.Base(uploadFilePath)

	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("[AWS_ACCESS_KEY]", "[AWS_SECRET_ACCESS_KEY]", "")),
	)
	if err != nil {
		color.Red("Failed setting aws config with error: %v", err.Error())
	}
	awsConfig.Region = AWS_REGION

	stat, err := os.Stat(uploadFilePath)
	if err != nil {
		color.Red("Couldn't stat file %v: %v", uploadFileName, err.Error())
	}
	file, err := os.Open(uploadFilePath)

	if err != nil {
		color.Red("Couldn't read file bytes %v: %v", uploadFileName, err.Error())
	}

	_, err = s3.NewFromConfig(awsConfig).PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(S3_BUCKET),
		Key:           aws.String(uploadFileName),
		Body:          file,
		ContentLength: stat.Size(),
	})

	file.Close()

	if err != nil {
		color.Red("File upload failed for %v: %v", uploadFileName, err.Error())
	}

	color.Green("File %v uploaded", uploadFileName)

}
