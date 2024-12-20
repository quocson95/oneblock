package api

import (
	"be/common"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	s3_credentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nickalie/go-webpbin"
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

func init() {
	uuid.EnableRandPool()
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

var DefaultS3Hepler = &S3Helper{}
var cachePreSign = make(map[string]*S3PreSign)

type S3Storage struct{}
type S3StorageResp struct {
	Url string `json:"url,omitempty"`
	Err error  `json:"err,omitempty"`
}

func (h *S3Storage) Handler(c *gin.RouterGroup) {
	c.GET("/", h.StorageFile)
	c.PUT("/", h.UploadStorageFile)
}

func (h *S3Storage) StorageFile(c *gin.Context) {
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

func (h *S3Storage) UploadStorageFile(c *gin.Context) {
	bucket := c.Query("bucket")
	name := c.Query("name")
	if len(bucket) == 0 {
		zap.L().Error("bucket missing")
		c.AbortWithStatus(http.StatusOK)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		zap.L().Error("read body failed")
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("read body failed")})
		return
	}
	bodyType := http.DetectContentType(body)
	originBodyLen := len(body)
	if bodyType == "image/png" || bodyType == "image/jpeg" {
		nameWithoutExt := name[:len(name)-len(filepath.Ext(name))]
		webpOut := &bytes.Buffer{}
		err = webpbin.NewCWebP().
			Quality(80).
			Input(bytes.NewReader(body)).
			Output(webpOut).
			Run()
		if err == nil {
			newName := nameWithoutExt + ".webp"
			newBody := webpOut.Bytes()
			zap.L().With(zap.String("origin name", name)).With(zap.Int("orgin size", originBodyLen)).
				With(zap.String("new name", newName)).With(zap.Int("new size", len(newBody))).With(zap.Int("save(%)", ((originBodyLen - len(newBody)) / (originBodyLen / 100)))).
				Info("convert image to webp ok")
			name = newName
			body = newBody
		} else {
			zap.L().Error("convert to webp")
			c.JSON(http.StatusOK, common.Mdx{Err: errors.New("convert to webp")})
			return
		}
	}
	preSign, err := DefaultS3Hepler.PreSign(http.MethodPut, bucket, name)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("presign put failed")
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("presign put failed")})
		return
	}
	client, err := http.NewRequest(http.MethodPut, preSign.Url, bytes.NewReader(body))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("init put failed")
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("init put failed")})
		return
	}
	client.Header.Set("Content-Type", c.ContentType())
	if len(client.Header.Get("Content-Type")) == 0 {
		client.Header.Set("Content-Type", http.DetectContentType(body))
	}
	client.ContentLength = int64(len(body))
	resp, err := common.DefaultHttpClient.Do(client)
	if err != nil {
		zap.L().With(zap.String("name", name)).With(zap.String("bucket", bucket)).With(zap.Error(err)).Error("upload failed")
		c.JSON(http.StatusBadRequest, &S3StorageResp{Err: errors.New("upload failed")})
		return
	}
	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		zap.L().With(zap.String("name", name)).With(zap.String("bucket", bucket)).With(zap.ByteString("body", bodyErr)).Error("upload failed")
		c.JSON(http.StatusBadRequest, &S3StorageResp{Err: errors.New("upload failed")})
		return
	}
	getPreSign, _ := url.Parse(fmt.Sprintf("https://%s/be/s3", c.Request.Host))
	getPreSign.RawQuery = ""
	query := url.Values{}
	query.Add("name", name)
	query.Add("bucket", bucket)
	getPreSign.RawQuery = query.Encode()
	c.JSON(http.StatusOK, &S3StorageResp{
		Url: getPreSign.String(),
	})
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
			Src:  preSign.Url,
			Size: *v.Size,
		}
		if v.Owner != nil {
			file.Filename = *v.Owner.DisplayName
		}
		resp.Files = append(resp.Files, file)
	}
	c.JSON(http.StatusOK, resp)
}
