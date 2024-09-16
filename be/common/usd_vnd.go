package common

type UsdVnd struct {
	Chart ChartUsdVnd `json:"chart"`
}

type ChartUsdVnd struct {
	Result []ResultUsdVnd `json:"result"`
	Error  interface{}    `json:"error"`
}

type ResultUsdVnd struct {
	Meta       Meta       `json:"meta"`
	Timestamp  []int64    `json:"timestamp"`
	Indicators Indicators `json:"indicators"`
}

type Indicators struct {
	Quote    []Quote    `json:"quote"`
	Adjclose []Adjclose `json:"adjclose"`
}

type Adjclose struct {
	Adjclose []*float64 `json:"adjclose"`
}

type Quote struct {
	Open   []*float64 `json:"open"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	High   []*float64 `json:"high"`
	Volume []*int64   `json:"volume"`
}

type Meta struct {
	Currency             string               `json:"currency"`
	Symbol               string               `json:"symbol"`
	ExchangeName         string               `json:"exchangeName"`
	FullExchangeName     string               `json:"fullExchangeName"`
	InstrumentType       string               `json:"instrumentType"`
	FirstTradeDate       int64                `json:"firstTradeDate"`
	RegularMarketTime    int64                `json:"regularMarketTime"`
	HasPrePostMarketData bool                 `json:"hasPrePostMarketData"`
	Gmtoffset            int64                `json:"gmtoffset"`
	Timezone             string               `json:"timezone"`
	ExchangeTimezoneName string               `json:"exchangeTimezoneName"`
	RegularMarketPrice   int64                `json:"regularMarketPrice"`
	FiftyTwoWeekHigh     int64                `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      int64                `json:"fiftyTwoWeekLow"`
	RegularMarketDayHigh int64                `json:"regularMarketDayHigh"`
	RegularMarketDayLow  int64                `json:"regularMarketDayLow"`
	RegularMarketVolume  int64                `json:"regularMarketVolume"`
	LongName             string               `json:"longName"`
	ShortName            string               `json:"shortName"`
	ChartPreviousClose   int64                `json:"chartPreviousClose"`
	PriceHint            int64                `json:"priceHint"`
	CurrentTradingPeriod CurrentTradingPeriod `json:"currentTradingPeriod"`
	DataGranularity      string               `json:"dataGranularity"`
	Range                string               `json:"range"`
	ValidRanges          []string             `json:"validRanges"`
}

type CurrentTradingPeriod struct {
	Pre     Post `json:"pre"`
	Regular Post `json:"regular"`
	Post    Post `json:"post"`
}

type Post struct {
	Timezone  string `json:"timezone"`
	End       int64  `json:"end"`
	Start     int64  `json:"start"`
	Gmtoffset int64  `json:"gmtoffset"`
}

type UsdVndBtc struct {
	Unix     []int     `json:"unix,omitempty"`
	VndUsd   []float64 `json:"vnd_usd,omitempty"`
	BtcPrice []float64 `json:"btc_price,omitempty"`
	Labels   []string  `json:"labels,omitempty"`
}
