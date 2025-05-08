package dashboard

import (
	"be/bot"
	"be/common"
	"be/security"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	payos "github.com/payOSHQ/payos-lib-golang"
	"github.com/skip2/go-qrcode"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var node *snowflake.Node
var tmpl *template.Template

func init() {
	node, _ = snowflake.NewNode(1)
	tmpl = template.Must(template.ParseFiles("static/payment_tele_template.html"))
}

func generateOrderCode() int64 {
	// Get the current ts in milliseconds
	ts := time.Now()
	// orderStr := fmt.Sprintf("%4d%2d%2d%2d%2d%2d", ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())

	// // Generate a random number to append (for added uniqueness)
	// rand.Seed(uint64(ts.UnixNano()))
	// randomNumber := rand.Intn(1000) // Generate a random number between 0 and 999

	// // Combine timestamp and random number to generate a unique order code
	// // We take the last 6 digits of the timestamp and add the random number
	// orderCode := int(ts.Unix()) + randomNumber

	return ts.UnixMilli()
}

type payOS struct {
	// clientId, apiKey, checksumKey string
	checksumKey []byte
}

func NewPayOS(clientId, apiKey, checksumKey string) *payOS {
	payos.Key(clientId, apiKey, checksumKey)
	return &payOS{
		// clientId:    clientId,
		// apiKey:      apiKey,
		checksumKey: []byte(checksumKey),
	}
}

func (p *payOS) Handler(g *echo.Group) {
	g.POST("/webhook", p.WebHook)
	g.GET("/", p.CreatePayment)
	g.GET("", p.CreatePayment)
	g.GET("/:id", p.GetPaymentInfo)
	g.GET("/qrcode/:id", p.PaymentQRImage)
}

func (p *payOS) WebHook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("read body failed")})
	}
	zap.L().With(zap.ByteString("body", body)).Info("payos hook")
	data := payos.WebhookType{}
	json.Unmarshal(body, &data)
	payOSHook, err := payos.VerifyPaymentWebhookData(data)
	if err != nil {
		zap.L().With(zap.ByteString("body", body)).Info("payos failed")
		return c.NoContent(http.StatusBadRequest)
	}
	payment := &common.Payment{}
	err = payment.FindByPaymentLinkID(payOSHook.PaymentLinkId)
	if err != nil {
		zap.L().With(zap.ByteString("body", body)).Info("find payment by PaymentLinkId failed")
		return c.NoContent(http.StatusOK)
	}
	payment.Currency = payOSHook.Currency
	payment.TransactionDate = payOSHook.TransactionDateTime
	if data.Success {
		payment.Status = common.StatusPaymentPaidSuccess
	} else {
		payment.Status = common.StatusPaymentPaidFailed
	}
	changes := make(map[string]interface{})
	changes["status"] = payment.Status
	changes["currency"] = payment.Currency
	changes["transaction_date"] = payment.TransactionDate
	payment.Update(changes)
	go func() {
		p.sendTeleNoti(payment)
		if payment.Status == common.StatusPaymentPaidSuccess {
			s := common.Subscribe{}
			err := s.GetByUser(payment.CustomerId)
			// not exist, create new
			extend, _ := time.ParseDuration(payment.Plan.DurationExtend)
			if err != nil || s.Id == 0 {
				s = common.Subscribe{
					CreatedAt:  time.Now(),
					ExpireUnix: time.Now().Add(extend).Unix(),
					UserId:     payment.CustomerId,
					PlanId:     payment.PlanId,
					Plan:       payment.Plan,
				}
				if err := s.Create(); err != nil {
					zap.L().With(zap.Int("customer id", payment.CustomerId)).With(zap.Error(err)).Error("insert subscribe failed")
				}
			}
			changes := make(map[string]interface{})
			// if diffirent plan, change plan
			if s.PlanId != payment.PlanId {
				changes["plan_id"] = payment.PlanId
			} else {
				changes["expire_unix"] = s.ExpireUnix + int64(extend.Seconds())
			}
			if err := s.Updates(changes); err != nil {
				zap.L().With(zap.Int("customer id", payment.CustomerId)).
					With(zap.Any("changes", changes)).With(zap.Error(err)).
					Error("update subscribe failed")
			}
		}
	}()
	return c.NoContent(http.StatusOK)
}

func (p *payOS) CreatePayment(c echo.Context) error {
	user := security.GetUserCtx(c)
	if user == nil {
		return c.NoContent(http.StatusBadRequest)
	}
	planId, _ := strconv.Atoi(c.QueryParam("id"))
	if planId <= 0 {
		return c.NoContent(http.StatusBadRequest)
	}
	plan := &common.Plan{
		Id: uint(planId),
	}
	err := plan.Get()
	if err != nil {
		zap.L().With(zap.Int("plan id", planId)).With(zap.Error(err)).Error("get plan failed")
		c.NoContent(http.StatusBadRequest)
	}
	query := make(url.Values)
	orderCode := generateOrderCode()
	query.Set("orderCode", strconv.FormatInt(orderCode, 10))
	query.Set("tk", quickSha1String([]byte(strconv.FormatInt(orderCode, 10)), p.checksumKey))
	body := payos.CheckoutRequestType{
		OrderCode: int64(orderCode),
		Amount:    plan.Price,
		Items: []payos.Item{
			{
				Name:     fmt.Sprintf("Oneblock Plan %d", planId),
				Price:    plan.Price,
				Quantity: 1,
			},
		},
		Description: fmt.Sprintf("Đơn hàng %d", orderCode),
		ReturnUrl:   "https://api.oneblock.vn/be/dashboard/payos/payment/return?" + query.Encode(),
		CancelUrl:   "https://api.oneblock.vn/be/dashboard/payos/payment/cancel?" + query.Encode(),
		BuyerEmail:  &user.Email,
	}
	cdCheckout := 20 * time.Minute
	expiredAt := int(time.Now().Add(cdCheckout).Add(1 * time.Minute).Unix())
	body.ExpiredAt = &expiredAt
	data, err := payos.CreatePaymentLink(body)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("create payment link failed")
		return c.NoContent(http.StatusBadRequest)
	}

	// q, err := qrcode.New(data.QRCode, qrcode.Medium)
	// if err != nil {
	// 	zap.L().With(zap.Error(err)).Error("generate qr code failed")
	// 	return c.NoContent(http.StatusBadRequest)
	// }
	payment := common.NewPayment(int(user.ID), data)
	payment.CustomerEmail = user.Email
	payment.QRCode = data.QRCode
	payment.CheckoutUrl = data.CheckoutUrl
	payment.Currency = data.Currency
	payment.Bin = data.Bin
	uri := c.Request().URL
	payment.QrLink = fmt.Sprintf("https://%s/%s/qrcode/%d", "api.oneblock.vn", strings.Trim(uri.Path, "/"), payment.OrderCode)
	payment.PlanId = int(plan.Id)
	payment.Plan = plan
	payment.CdCheckoutBySec = int64(cdCheckout.Seconds())
	payment.Insert()
	return c.JSON(http.StatusOK, payment)
}

func (p *payOS) PaymentQRImage(c echo.Context) error {
	// user := security.GetUserCtx(c)
	orderId := c.Param("id")
	orderCode, _ := strconv.Atoi(orderId)
	payment := &common.Payment{}
	payment.FindByOrderCode(0, int64(orderCode))
	q, err := qrcode.New(payment.QRCode, qrcode.Medium)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("generate qr code failed")
		return c.NoContent(http.StatusBadRequest)
	}
	res := c.Response()
	res.Header().Set("Content-Type", "image/png")
	res.WriteHeader(http.StatusOK)
	q.Write(256, res.Writer)
	return nil
}

func (p *payOS) GetPaymentInfo(c echo.Context) error {
	user := security.GetUserCtx(c)
	orderId := c.Param("id")
	orderCode, _ := strconv.Atoi(orderId)
	payment := &common.Payment{}
	payment.FindByOrderCode(int(user.ID), int64(orderCode))
	payment.CdCheckoutBySec -= time.Now().Unix() - payment.CreatedAt.Unix()
	return c.JSON(http.StatusOK, payment)
}

func (p *payOS) sendTeleNoti(payment *common.Payment) {
	// msg:=fmt.Sprintf("")
	buff := bytes.Buffer{}
	err := tmpl.Execute(&buff, payment)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("parse template send noti tele failed")
		return
	}
	bot.TeleBot.SendMsgChannel(buff.String())
}

func verifyChecksumDataHook(data []byte, checksumKey string) bool {
	webhookData := make(map[string]interface{})
	json.Unmarshal(data, &webhookData)
	signature := webhookData["signature"].(string)
	return isValidData(webhookData, signature, checksumKey)
}

// Sorts an object by its keys (as a map) in a similar manner as the JavaScript code
func sortMapByKey(m map[string]interface{}) map[string]interface{} {
	sortedMap := make(map[string]interface{})
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		sortedMap[key] = m[key]
	}
	return sortedMap
}

// Converts an object to a query string
func convertObjToQueryStr(object map[string]interface{}) string {
	var parts []string
	for key, value := range object {
		if value != nil {
			// Handle nested array
			if arr, ok := value.([]interface{}); ok {
				// Sort the array items if they are objects
				var sortedArr []string
				for _, val := range arr {
					// If the value is a map, sort its keys
					if subMap, ok := val.(map[string]interface{}); ok {
						sortedMap := sortMapByKey(subMap)
						sortedArr = append(sortedArr, convertObjToQueryStr(sortedMap))
					}
				}
				// Join all the sorted items in the array
				parts = append(parts, fmt.Sprintf("%s=%s", key, "["+strings.Join(sortedArr, ",")+"]"))
			} else {
				// Normal string value
				parts = append(parts, fmt.Sprintf("%s=%v", key, value))
			}
		}
	}
	return strings.Join(parts, "&")
}

// Verifies the data by calculating the HMAC and comparing it to the given signature
func isValidData(data map[string]interface{}, currentSignature, checksumKey string) bool {
	// Sort the data by keys
	sortedDataByKey := sortMapByKey(data)

	// Convert the sorted data to a query string
	dataQueryStr := convertObjToQueryStr(sortedDataByKey)

	// Calculate the HMAC-SHA256 signature using the checksum key
	// hmac := hmac.New(sha256.New, []byte(checksumKey))
	// hmac.Write([]byte(dataQueryStr))
	// calculatedSignature := hex.EncodeToString(hmac.Sum(nil))
	calculatedSignature := quickSha1String([]byte(dataQueryStr), []byte(checksumKey))

	// Return whether the calculated signature matches the current signature
	return calculatedSignature == currentSignature
}

func quickSha1String(dataQueryStr []byte, checksumKey []byte) string {
	hmac := hmac.New(sha256.New, checksumKey)
	hmac.Write(dataQueryStr)
	calculatedSignature := hex.EncodeToString(hmac.Sum(nil))
	return calculatedSignature
}
