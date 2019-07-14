package router

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
)

var (
	accessKey string
	secretKey string
)

func init() {
	viper.BindEnv("A")
	accessKey = viper.GetString("A")
	viper.BindEnv("B")
	secretKey = viper.GetString("B")
}

func uploadFile(w http.ResponseWriter, r *http.Request, uuid string) (string, error) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("image")
	if err != nil {
		return "", err
	}
	defer file.Close()
	f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, file)
	return s3UploadFile(f, handler.Filename, uuid)
}

func s3UploadFile(f *os.File, filename string, uuid string) (string, error) {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(
			accessKey, // id
			secretKey, // secret
			""), // token can be left blank for now
	}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("doughyou-images"),
		Key:    aws.String(fmt.Sprintf("foods/%s/%s", uuid, filename)),
		Body:   f,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	return aws.StringValue(&result.Location), nil
}
