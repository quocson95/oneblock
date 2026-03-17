package common

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	s3_credentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3StorageResp struct {
	Url string `json:"url,omitempty"`
	Err error  `json:"err,omitempty"`
}

var DefaultS3Hepler = &S3Helper{}
var Idrivee2S3Helper = &S3Helper{}

type S3PreSign struct {
	Method string `json:"method,omitempty"`
	Url    string `json:"url,omitempty"`
	Expire int64  `json:"expire,omitempty"`
}
type S3Helper struct {
	svc *s3.S3
}

func (s *S3Helper) Init(endpoint, region, accessKey, secretKey string) error {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: s3_credentials.NewStaticCredentials(
			accessKey, secretKey, "",
		),
	})
	if err != nil {
		return err
	}
	s.svc = s3.New(sess)
	return nil
}
func (s *S3Helper) SVC() *s3.S3 {
	return s.svc
}
func (s *S3Helper) PreSign(method string, bucket string, nameFile string) (*S3PreSign, error) {
	var req *request.Request
	if method == http.MethodGet {
		req, _ = s.svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(nameFile),
		})
	} else {
		req, _ = s.svc.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(nameFile),
		})
	}

	expire := 23 * time.Hour
	preSign, err := req.Presign(expire)
	if err != nil {
		return &S3PreSign{}, err
	}
	return &S3PreSign{
		Method: method,
		Url:    preSign,
		Expire: time.Now().Add(expire).Add(-10 * time.Minute).Unix(),
	}, nil
}
