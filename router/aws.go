package router

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
	"mime/multipart"
	"net/http"
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

func uploadFile(w http.ResponseWriter, r *http.Request, uuid string) (string, []string, error) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("image")
	if err != nil {
		return "", nil, err
	}
	defer file.Close()
	filename := handler.Filename
	url, err := s3UploadFile(file, filename, uuid)
	if err != nil {
		return "", nil, err
	}
	labels, err := rekognizeImage(uuid, filename)
	if err != nil {
		return "", nil, err
	}

	tags := []string{}
	for _, l := range labels {
		tags = append(tags, aws.StringValue(l.Name))
	}
	return url, tags, nil
}

func s3UploadFile(f multipart.File, filename string, uuid string) (string, error) {
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

func rekognizeImage(uuid string, filename string) ([]*rekognition.Label, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(
			accessKey, // id
			secretKey, // secret
			""), // token can be left blank for now
	}))
	svc := rekognition.New(sess)
	out, err := svc.DetectLabels(&rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String("doughyou-images"),
				Name:   aws.String(fmt.Sprintf("foods/%s/%s", uuid, filename)),
			},
		},
		MaxLabels:     aws.Int64(5),
		MinConfidence: aws.Float64(95.5),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to rekognize image, %v", err.Error())
	}

	return out.Labels, nil
}
