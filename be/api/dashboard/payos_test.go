package dashboard

import "testing"

func Test_verifyChecksumDataHook(t *testing.T) {
	name := `Test_verifyChecksumDataHook`
	jsonStr := `{
		"code": "00",
		"desc": "success",
		"success": true,
		"data": {
			"orderCode": 123,
			"amount": 3000,
			"description": "VQRIO123",
			"accountNumber": "12345678",
			"reference": "TF230204212323",
			"transactionDateTime": "2023-02-04 18:25:00",
			"currency": "VND",
			"paymentLinkId": "124c33293c43417ab7879e14c8d9eb18",
			"counterAccountBankId": "",
			"counterAccountBankName": "",
			"counterAccountName": "",
			"counterAccountNumber": "",
			"virtualAccountName": "",
			"virtualAccountNumber": ""
		},
		"signature": "412e915d2871504ed31be63c8f62a149a4410d34c4c42affc9006ef9917eaa03"
	}`
	t.Run(name, func(t *testing.T) {
		verifyChecksumDataHook([]byte(jsonStr), "1a54716c8f0efb2744fb28b6e38b25da7f67a925d98bc1c18bd8faaecadd7675")
	})
}
