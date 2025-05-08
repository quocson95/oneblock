package api

import (
	"be/common"
	"be/job"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nickalie/go-webpbin"
	"go.uber.org/zap"
)

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

type S3Storage struct{}

var wlBucket = map[string]struct{}{"cms-images": {}, common.DefaultBucketMdx.String(): {}}

func (h *S3Storage) Handler(c *echo.Group) {
	c.GET("", h.StorageFile)
	c.GET("/", h.StorageFile)
	c.PUT("/", h.UploadStorageFile)
}

func (h *S3Storage) StorageFile(c echo.Context) error {
	bucket := c.QueryParam("bucket")
	name := c.QueryParam("name")
	noCache := c.QueryParam("noCache") == "true"
	if _, exsit := wlBucket[bucket]; !exsit {
		return echo.NewHTTPError(http.StatusOK)

	}
	if len(name) == 0 {
		return h.StorageListObject(c)

	}
	key := bucket + name
	if w, _ := strconv.Atoi(c.QueryParam("w")); w > 0 {
		keyW := fmt.Sprintf("%s_w%d", key, w)
		image, exist := common.CacheDataPool.Get(keyW)
		if exist {
			header := image.Header
			header["Cache-Control"] = "public, max-age=86400"
			header["Expires"] = image.InvalidAt.Format(http.TimeFormat)
			header["Last-Modified"] = image.CreateAt.Format(http.TimeFormat)
			header["Content-Encoding"] = image.ContentEncoding
			resp := c.Response()
			for k, v := range header {
				resp.Header().Set(k, v)
			}
			c.SetResponse(resp)
			return c.Blob(http.StatusOK, image.Mime, image.Data)
		}
	}
	v, exist := common.CacheDataPool.Get(key)
	var preSign *common.S3PreSign
	var err error
	if noCache || !exist {
		zap.L().With(zap.String("bucket", bucket)).With(zap.String("name", name)).Info("not found cache image")
		preSign, err = common.DefaultS3Hepler.PreSign(http.MethodGet, bucket, name)
		if err != nil {
			zap.L().With(zap.Error(err)).With(zap.String("bucket", bucket)).With(zap.String("name", name)).Error("presign failed")
			return echo.NewHTTPError(http.StatusBadRequest, "presign failed")

		}
		resp, cleanup, err := common.QuickGetHttp(http.MethodGet, preSign.Url, nil)
		defer cleanup()
		if err != nil {
			zap.L().With(zap.Error(err)).With(zap.String("bucket", bucket)).With(zap.String("name", name)).Error("get content presign failed")
			return echo.NewHTTPError(http.StatusBadRequest, "presign failed")

		}
		image := common.CacheData{
			ContentLength: resp.ContentLength,
			Mime:          resp.Header.Get("content-type"),
			Header:        make(map[string]string),
			InvalidAt:     time.Now().Add(30 * 24 * time.Hour),
			CreateAt:      time.Now(),
		}
		for k, v := range resp.Header {
			if len(v) == 0 {
				continue
			}
			image.Header[k] = v[0]
		}
		image.Data, _ = io.ReadAll(resp.Body)
		common.CacheDataPool.Add(key, image)
		v = image
		exist = true
		if w, _ := strconv.Atoi(c.QueryParam("w")); w > 0 {
			key = fmt.Sprintf("%s_w%d", key, w)
			common.CacheDataPool.Add(key, image)
			job.AddCompressImage(key, w)
		}
	}

	if !exist {
		return c.Redirect(http.StatusFound, preSign.Url)
	}

	image := v
	header := image.Header
	header["Cache-Control"] = "public, max-age=2592000"
	header["Expires"] = image.InvalidAt.Format(http.TimeFormat)
	header["Last-Modified"] = image.CreateAt.Format(http.TimeFormat)
	header["Content-Encoding"] = image.ContentEncoding
	// c.DataFromReader(http.StatusOK, int64(image.ContentLength), image.Mime, bytes.NewBuffer(image.Data), header)
	resp := c.Response()
	for k, v := range header {
		resp.Header().Set(k, v)
	}
	c.SetResponse(resp)
	return c.Blob(http.StatusOK, image.Mime, image.Data)
}

func (h *S3Storage) UploadStorageFile(c echo.Context) error {
	bucket := c.QueryParam("bucket")
	name := c.QueryParam("name")
	if len(bucket) == 0 {
		zap.L().Error("bucket missing")
		return c.NoContent(http.StatusOK)
	}
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		zap.L().Error("read body failed")
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("read body failed")})
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
	preSign, err := common.DefaultS3Hepler.PreSign(http.MethodPut, bucket, name)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("presign put failed")
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("presign put failed")})
	}
	client, err := http.NewRequest(http.MethodPut, preSign.Url, bytes.NewReader(body))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("init put failed")
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("init put failed")})
	}
	client.Header.Set("Content-Type", c.Request().Header.Get("Content-Type"))
	if len(client.Header.Get("Content-Type")) == 0 {
		client.Header.Set("Content-Type", http.DetectContentType(body))
	}
	client.ContentLength = int64(len(body))
	resp, err := common.DefaultHttpClient.Do(client)
	if err != nil {
		zap.L().With(zap.String("name", name)).With(zap.String("bucket", bucket)).With(zap.Error(err)).Error("upload failed")
		return c.JSON(http.StatusBadRequest, &common.S3StorageResp{Err: errors.New("upload failed")})
	}
	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		zap.L().With(zap.String("name", name)).With(zap.String("bucket", bucket)).With(zap.ByteString("body", bodyErr)).Error("upload failed")
		return c.JSON(http.StatusBadRequest, &common.S3StorageResp{Err: errors.New("upload failed")})
	}
	key := bucket + name
	common.CacheDataPool.Remove(key)
	s3Sync := &common.S3ObjectSync{
		Name:     name,
		Bucket:   bucket,
		Source_1: common.SourceS3CloudFy,
	}
	if err := s3Sync.Insert(); err != nil {
		zap.L().With(zap.Error(err)).Error("save s3 sync object failed")
	}
	getPreSign, _ := url.Parse(fmt.Sprintf("https://%s/be/s3", c.Request().Host))
	getPreSign.RawQuery = ""
	query := url.Values{}
	query.Add("name", name)
	query.Add("bucket", bucket)
	getPreSign.RawQuery = query.Encode()
	return c.JSON(http.StatusOK, &common.S3StorageResp{
		Url: getPreSign.String(),
	})
}

func (h *S3Storage) StorageListObject(c echo.Context) error {
	offset := 0
	if v, _ := strconv.Atoi(c.QueryParam("offset")); v > 0 {
		offset = v
	}
	limit := 10
	if v, _ := strconv.Atoi(c.QueryParam("limit")); v > 0 {
		limit = v
	}
	bucket := c.QueryParam("bucket")

	ml, err := common.GetS3ObjectsSync(bucket, offset, limit)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get list s3 object failed")
		return c.NoContent(http.StatusBadRequest)
	}
	for idx, v := range ml {
		// preSignGet, _ := common.DefaultS3Hepler.PreSign(http.MethodGet, v.Bucket, v.Name)
		v.PresignUrl = fmt.Sprintf("https://api.oneblock.vn/be/s3/?bucket=%s&name=%s", v.Bucket, v.Name)
		ml[idx] = v
	}
	return c.JSON(http.StatusOK, ml)
}
