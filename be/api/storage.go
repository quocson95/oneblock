package api

import (
	"be/cache"
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
		h.StorageListObject(c)
		return
	}
	key := bucket + name
	v, exist := cache.CacheData.Load(key)
	var preSign *S3PreSign
	var err error
	if !exist {
		zap.L().With(zap.String("bucket", bucket)).With(zap.String("name", name)).Info("not found cache image")
		preSign, err = DefaultS3Hepler.PreSign(http.MethodGet, bucket, name)
		if err != nil {
			zap.L().With(zap.Error(err)).With(zap.String("bucket", bucket)).With(zap.String("name", name)).Error("presign failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, "presign failed")
			return
		}
		resp, cleanup, err := common.QuickGetHttp(http.MethodGet, preSign.Url, nil)
		defer cleanup()
		if err != nil {
			zap.L().With(zap.Error(err)).With(zap.String("bucket", bucket)).With(zap.String("name", name)).Error("get content presign failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, "presign failed")
			return
		}
		image := cache.ImageData{
			ContentLength: resp.ContentLength,
			Mime:          resp.Header.Get("content-type"),
			Header:        make(map[string]string),
			InvalidAt:     time.Now().Add(24 * time.Hour),
			CreateAt:      time.Now(),
		}
		for k, v := range resp.Header {
			if len(v) == 0 {
				continue
			}
			image.Header[k] = v[0]
		}
		image.Data, _ = io.ReadAll(resp.Body)
		cache.CacheData.Store(key, image)
		v = image
		exist = true
	}

	if !exist {
		c.Redirect(http.StatusFound, preSign.Url)
		return
	}
	image := v.(cache.ImageData)
	header := image.Header
	header["Cache-Control"] = "public, max-age=86400"
	header["Expires"] = image.InvalidAt.Format(http.TimeFormat)
	header["Last-Modified"] = image.CreateAt.Format(http.TimeFormat)
	c.DataFromReader(http.StatusOK, int64(image.ContentLength), image.Mime, bytes.NewBuffer(image.Data), header)
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
		err = webpbin.NewCWebP().Quality(80).Input(bytes.NewReader(body)).Output(webpOut).Run()
		if err == nil {
			newName := nameWithoutExt + ".webp"
			newBody := webpOut.Bytes()
			zap.L().With(zap.String("origin name", name)).With(zap.Int("orgin size", originBodyLen)).
				With(zap.String("new name", newName)).With(zap.Int("new size", len(newBody))).With(zap.Int("save(%)", ((originBodyLen - len(newBody)) / (originBodyLen / 100)))).
				Info("convert image to webp ok")
			name = newName
			body = newBody
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
	s3Sync := &common.S3ObjectSync{
		Name:     name,
		Bucket:   bucket,
		Source_1: common.SourceS3CloudFy,
	}
	if err := s3Sync.Insert(); err != nil {
		zap.L().With(zap.Error(err)).Error("save s3 sync object failed")
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

func (h *S3Storage) StorageListObject(c *gin.Context) {
	offset := c.GetInt("offset")
	limit := c.GetInt("limit")
	bucket := c.GetString("bucket")
	if limit <= 0 {
		limit = 10
	}
	ml, err := common.GetS3ObjectsSync(bucket, offset, limit)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get list s3 object failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	for idx, v := range ml {
		preSignGet, _ := DefaultS3Hepler.PreSign(http.MethodGet, v.Bucket, v.Name)
		v.PresignUrl = preSignGet.Url
		ml[idx] = v
	}
	c.JSON(http.StatusOK, ml)
}
