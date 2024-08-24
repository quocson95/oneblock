package common

import (
	"encoding/json"
	"fmt"
	"time"
)

type HolderBtc struct {
	Counts []HolderBtcElement `json:"counts"`
	Prices []HolderBtcElement `json:"prices"`
}

type HolderBtcElement struct {
	T int64   `json:"t"`
	V float64 `json:"v"`
}

func (h *HolderBtc) Load(fromTime time.Time) error {
	fn := func(id DataRawID) ([]HolderBtcElement, error) {
		ml := make([]HolderBtcElement, 0)
		err := Load(id, func(data []byte) error {
			return json.Unmarshal(data, &ml)
		})
		res := make([]HolderBtcElement, 0)
		fromTimeUnix := fromTime.Unix()
		for _, m := range ml {
			if m.T < fromTimeUnix {
				continue
			}
			res = append(res, m)
		}
		if err != nil {
			return nil, fmt.Errorf("load %s error: %w", id, err)
		}
		return res, nil
	}
	var err error
	if h.Counts, err = fn(HolderBtcCountId); err != nil {
		return err
	}
	if h.Prices, err = fn(HolderBtcPriceId); err != nil {
		return err
	}
	return nil
}
