package entity

import (
	"backup-x/util"
	"errors"
	"log"
	"os"
	"strings"

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

var ErrS3Empty = errors.New("s3 config is empty")

func (s3Config S3Config) CheckNotEmpty() bool {
	return s3Config.Endpoint != "" && s3Config.AccessKey != "" &&
		s3Config.SecretKey != "" && s3Config.BucketName != ""
}

func (s3Config S3Config) getSession() (*session.Session, error) {

	if !s3Config.CheckNotEmpty() {
		return nil, ErrS3Empty
	}

	conf, err := GetConfigCache()
	if err != nil {
		return nil, err
	}
	secretKey, err := util.DecryptByEncryptKey(conf.EncryptKey, s3Config.SecretKey)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewStaticCredentials(s3Config.AccessKey, secretKey, "")
	_, err = creds.Get()
	if err != nil {
		log.Println(err)
	}

	region := "cn-north-1"
	if strings.HasSuffix(s3Config.Endpoint, "amazonaws.com") {
		sp := strings.Split(s3Config.Endpoint, ".")
		if len(sp) > 1 {
			region = sp[1]
		}
	}
	config := &aws.Config{
		Region:           aws.String(region),
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
		if err != ErrS3Empty {
			log.Printf("创建对象存储会话失败, ERR: %s\n", err)
		}
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

// UploadFile 上传
func (s3Config S3Config) UploadFile(fileName string) {
	mySession, err := s3Config.getSession()
	if err != nil {
		if err != ErrS3Empty {
			log.Printf("创建对象存储会话失败, ERR: %s\n", err)
		}
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

// ListFiles 列出文件
func (s3Config S3Config) ListFiles(projectPath string) (fileNames []string, err error) {
	mySession, err := s3Config.getSession()
	if err != nil {
		if err != ErrS3Empty {
			log.Printf("创建对象存储会话失败, ERR: %s\n", err)
		}
		return nil, err
	}

	svc := s3.New(mySession)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(s3Config.BucketName),
		Prefix: aws.String(projectPath),
	}
	resp, err := svc.ListObjects(params)
	if err != nil {
		return nil, err
	}

	for _, item := range resp.Contents {
		fileNames = append(fileNames, *item.Key)
	}

	return fileNames, err
}

// DeleteFile 删除文件
func (s3Config S3Config) DeleteFile(filePath string) error {
	mySession, err := s3Config.getSession()
	if err != nil {
		if err != ErrS3Empty {
			log.Printf("创建对象存储会话失败, ERR: %s\n", err)
		}
		return err
	}

	svc := s3.New(mySession)
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(s3Config.BucketName), Key: aws.String(filePath)})
	if err != nil {
		return err
	}

	return svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s3Config.BucketName),
		Key:    aws.String(filePath),
	})
}
