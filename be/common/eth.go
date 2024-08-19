package common

type EthGas struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  Result `json:"result"`
}

type Result struct {
	LastBlock       string `json:"LastBlock"`
	SafeGasPrice    string `json:"SafeGasPrice"`
	ProposeGasPrice string `json:"ProposeGasPrice"`
	FastGasPrice    string `json:"FastGasPrice"`
	SuggestBaseFee  string `json:"suggestBaseFee"`
	GasUsedRatio    string `json:"gasUsedRatio"`
}

type BtcEth struct {
	BtcCurrentPrice
	EthGas
}
