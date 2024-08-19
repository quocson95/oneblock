package craw

import (
	"be/common"
	"be/config"
	"encoding/json"
	"io"
	"net/http"
)

func EthGas() (*common.EthGas, error) {
	url := "https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey=" + config.GetConfig().ApiEthKey
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	ethGas := &common.EthGas{}
	json.Unmarshal(body, ethGas)
	return ethGas, nil
}
