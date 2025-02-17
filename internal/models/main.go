package models

type Trade struct {
	ExchangeName string `json:"ff,omitempty"`
	FirstPrice   string `json:"f2,omitempty"`
	LastPrice    string `json:"f3,omitempty"`
	EventType    string `json:"e"`
	EventTime    int64  `json:"E"`
	Time         int64  `json:"T"`
	TradeID      int64  `json:"t"`
	Symbol       string `json:"s"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	Buyer        bool   `json:"m"`
}

type OrderBook struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Time      int64  `json:"T"`
	UpdateID  int64  `json:"u"`
	Symbol    string `json:"s"`
	BidPrice  string `json:"b"`
	BidQty    string `json:"B"`
	AskPrice  string `json:"a"`
	AskQty    string `json:"A"`
}

type Liquidation struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Order     struct {
		Price    string `json:"p"`
		Quantity string `json:"q"`
		Side     string `json:"S"`
	} `json:"o"`
}
