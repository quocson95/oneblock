package job

import (
	"be/common"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type CrawlDataTrading struct{}

func (c *CrawlDataTrading) BtcGoldCore() error {
	url := "https://newhedge.io/api/v1/analytics/session/68"
	method := "POST"

	// client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("x-csrf-token", "Ve11IFU1vBralFO7LQziEVUDBu6stzZo9u-9i_ukkdmSFrMSYtJSqoHePYCH5_MthFpOtbMj4aDEr9KDXtZWgA")
	req.Header.Add("Cookie", "_session_id=Pmu7yyD%2F4ttsleffjSrf3OIHDeC1zVE0OvIfeoMIqBCHdjhUGUjoJt0SHjnswswyRjPBPm3s685KHye0P0vVZ1rdFoWWbjJwAbh9aE06obzuv8L3dACr6FpbbIlRueLAf8td%2B3G22LEqtWLbdvCDAOERaVHQYkyBvcnveULudnhGqBnWTDaqWY7aLkGe28Lf7iF8lWlBt9QG76VmaEPmLIwjYjQpTFtqR99ltkMAJYBPvnC7Q3CDyJ1gVD7e9MYK7t%2BztA3WmjFizpdeLGO%2B6yi5wW0zyfBFcyxl6vL04EUoeWOVrWPrE5M%3D--bJiC%2Bh1i%2FPguB48B--Y1R7etJX0d1gI2VnLInNEA%3D%3D")

	res, err := common.DefaultHttpClient.Do(req)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("craw btc gold core failed")
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("read body craw btc gold core failed")
		return err
	}
	data, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("decode body craw btc gold core failed")
		return err
	}
	logEvent := common.CrawlLog{
		EventId: common.CrawlLogEventIdDataTradingBtcGold,
	}
	if err := logEvent.Insert(); err != nil {
		zap.L().With(zap.Error(err)).Error("insert log craw failed")
	}
	name := "btc_gold.json"
	return c.upload(name, data)
}

func (c CrawlDataTrading) Sp500() error {
	url := "https://api.investing.com/api/financialdata/166/historical/chart/?interval=P1M&pointscount=160"
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("create req failed")
		return err
	}
	req.Header.Add("domain-id", "vn")
	req.Header.Add("origin", "https://vn.investing.com")
	req.Header.Add("referer", "https://vn.investing.com/")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 YaBrowser/24.7.0.0 Safari/537.36")
	res, err := common.DefaultHttpClient.Do(req)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("craw btc gold core failed")
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("read body craw btc gold core failed")
		return err
	}
	logEvent := common.CrawlLog{
		EventId: common.CrawlLogEventIdDataTradingSP500,
	}
	if err := logEvent.Insert(); err != nil {
		zap.L().With(zap.Error(err)).Error("insert log craw failed")
	}
	// fmt.Println(string(data))
	name := "sp500.json"
	return c.upload(name, data)
}

func (c *CrawlDataTrading) upload(name string, data []byte) error {
	bucket := common.DefaultBucketCrawl.String()
	preSigUrl, _ := common.DefaultS3Hepler.PreSign(http.MethodPut, bucket, name)
	req, err := common.NewHttpRequest(http.MethodPut, preSigUrl.Url, data)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("create req upload failed")
		return err
	}
	resp, err := common.DefaultHttpClient.Do(req)
	if err != nil {
		zap.L().With(zap.String("name", name)).With(zap.String("bucket", bucket)).With(zap.Error(err)).Error("upload failed")
		return err
	}
	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		zap.L().With(zap.String("name", name)).With(zap.String("bucket", bucket)).With(zap.ByteString("body", bodyErr)).Error("upload failed")
		return err
	}
	s3Sync := &common.S3ObjectSync{
		Name:     name,
		Bucket:   bucket,
		Source_1: common.SourceS3CloudFy,
	}
	if err := s3Sync.Insert(); err != nil {
		zap.L().With(zap.String("name", s3Sync.Name)).With(zap.String("bucket", s3Sync.Bucket)).With(zap.Error(err)).Error("insert s3 sync failed")
	}
	return nil
}
