package api

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	s3_credentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

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

func (s *S3Helper) PreSign(method string, bucket string, nameFile string) (*S3PreSign, error) {
	req, _ := s.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(nameFile),
	})
	expire := 23 * time.Hour
	preSign, err := req.Presign(expire)
	if err != nil {
		return nil, err
	}
	return &S3PreSign{
		Method: method,
		Url:    preSign,
		Expire: time.Now().Add(expire).Add(-10 * time.Minute).Unix(),
	}, nil
}

var DefaultS3Hepler = &S3Helper{}
var cachePreSign = make(map[string]*S3PreSign)

func StorageFile(c echo.Context) error {
	bucket := c.QueryParam("bucket")
	name := c.QueryParam("name")
	key := bucket + name
	v, exist := cachePreSign[key]
	if exist && v.Expire > time.Now().Unix() {
		return c.Redirect(http.StatusFound, v.Url)
	}
	preSign, err := DefaultS3Hepler.PreSign(http.MethodGet, bucket, name)
	if err != nil {
		zap.L().With(zap.Error(err)).With(zap.String("bucket", bucket)).With(zap.String("name", name)).Error("presign failed")
		return echo.NewHTTPError(http.StatusBadRequest, "presign failed")
	}
	cachePreSign[key] = preSign
	return c.Redirect(http.StatusFound, preSign.Url)
}
