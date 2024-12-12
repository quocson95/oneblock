package api

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	s3_credentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
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

type S3Dirs struct {
	Files []File `json:"files"`
	// Directories []interface{} `json:"directories"`
}

type File struct {
	Src      string `json:"src"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
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

func StorageFile(c *gin.Context) {
	bucket := c.Query("bucket")
	name := c.Query("name")
	if len(name) == 0 {
		listS3(c)
		return
	}
	key := bucket + name
	v, exist := cachePreSign[key]
	if exist && v.Expire > time.Now().Unix() {
		c.Redirect(http.StatusFound, v.Url)
		return
	}
	preSign, err := DefaultS3Hepler.PreSign(http.MethodGet, bucket, name)
	if err != nil {
		zap.L().With(zap.Error(err)).With(zap.String("bucket", bucket)).With(zap.String("name", name)).Error("presign failed")
		c.AbortWithStatusJSON(http.StatusBadRequest, "presign failed")
	}
	cachePreSign[key] = preSign
	c.Redirect(http.StatusFound, preSign.Url)
}

func listS3(c *gin.Context) {
	svc := DefaultS3Hepler.SVC()
	bucket := c.Query("bucket")
	lst, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		c.JSON(http.StatusOK, "")
	}
	resp := &S3Dirs{
		Files: make([]File, 0),
	}
	for _, v := range lst.Contents {
		preSign, _ := DefaultS3Hepler.PreSign(http.MethodGet, bucket, *v.Key)

		file := File{
			Src:      preSign.Url,
			Filename: *v.Owner.DisplayName,
			Size:     *v.Size,
		}
		resp.Files = append(resp.Files, file)
	}
	c.JSON(http.StatusOK, resp)
	return
}
