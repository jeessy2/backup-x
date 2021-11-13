package entity

import (
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Config S3Config
type S3Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
}

func (s3Config S3Config) checkNotEmpty() bool {
	return s3Config.Endpoint != "" && s3Config.AccessKey != "" &&
		s3Config.SecretKey != "" && s3Config.BucketName != ""
}

func (s3Config S3Config) getSession() (*session.Session, error) {

	if !s3Config.checkNotEmpty() {
		return nil, errors.New("s3 config is empty")
	}

	creds := credentials.NewStaticCredentials(s3Config.AccessKey, s3Config.SecretKey, "")
	_, err := creds.Get()
	if err != nil {
		log.Println(err)
	}

	config := &aws.Config{
		Region:           aws.String("cn-north-1"),
		Endpoint:         aws.String(s3Config.Endpoint),
		DisableSSL:       aws.Bool(false),
		Credentials:      creds,
		S3ForcePathStyle: aws.Bool(true),
	}

	mySession, err := session.NewSession(config)
	return mySession, err
}

func (s3Config S3Config) CreateBucketIfNotExist() {
	mySession, err := s3Config.getSession()
	if err != nil {
		return
	}
	client := s3.New(mySession)

	head := &s3.HeadBucketInput{
		Bucket: aws.String(s3Config.BucketName),
	}
	_, err = client.HeadBucket(head)

	if err != nil {
		create := &s3.CreateBucketInput{
			Bucket: aws.String(s3Config.BucketName),
		}
		_, err = client.CreateBucket(create)
		if err != nil {
			log.Printf("创建bucket: %s 失败, ERR: %s\n", s3Config.BucketName, err)
		} else {
			log.Printf("创建bucket: %s 成功\n", s3Config.BucketName)
		}
	}
}

func (s3Config S3Config) UploadFile(fileName string) {
	mySession, err := s3Config.getSession()
	if err != nil {
		return
	}

	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	uploader := s3manager.NewUploader(mySession)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Config.BucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		log.Printf("%s 上传到对象存储失败. ERR: %s \n", fileName, err)
	} else {
		log.Printf("%s 上传到对象存储成功\n", fileName)
	}
}
