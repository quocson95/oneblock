package dashboard

// import (
// 	"be/common"
// 	"be/security"
// 	"net/http"

// 	"github.com/ducnpdev/vietqr"
// 	"github.com/labstack/echo/v4"
// 	"github.com/skip2/go-qrcode"
// 	"go.uber.org/zap"
// )

// type PaymentController struct{}

// func (p *PaymentController) Handler(g *echo.Group) {
// 	g.GET("/", p.CreateQRPayment)
// 	g.GET("", p.CreateQRPayment)
// }

// const AccIDTech = "970407"

// func (p *PaymentController) CreateQRPayment(c echo.Context) error {
// 	user := security.GetUserCtx(c)
// 	if user == nil {
// 		return c.NoContent(http.StatusBadRequest)
// 	}
// 	payment := common.NewPayment(int(user.ID))
// 	payment.Amount = "5000"
// 	payment.BankAccountNo = "9482345678"
// 	payment.AcqID = AccIDTech
// 	payment.Unit = "vnd"
// 	content := vietqr.GenerateViQR(vietqr.RequestGenerateViQR{
// 		MerchantAccountInformation: vietqr.MerchantAccountInformation{
// 			AccountNo: payment.BankAccountNo,
// 			AcqID:     payment.AcqID,
// 		},
// 		TransactionAmount: payment.Amount,
// 		AdditionalDataFieldTemplate: vietqr.AdditionalDataFieldTemplate{
// 			Description: payment.TransactionId,
// 		},
// 	})
// 	png, err := qrcode.Encode(content, qrcode.Medium, 256)
// 	if err != nil {
// 		zap.L().With(zap.String("content", content)).With(zap.Error(err)).Error("generate qr code failed")
// 		return c.NoContent(http.StatusBadRequest)
// 	}
// 	payment.Insert()
// 	return c.Blob(http.StatusOK, "image/png", png)
// }
