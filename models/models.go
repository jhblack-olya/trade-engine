/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package models

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Used to indicate the direction of an order or transaction: buy, sellx
type Side string

var CommonError map[string]string
var RedisErrCh chan error
var MysqlErrCh chan error
var KafkaErrCh chan error
var Trigger chan int64

func NewSideFromString(s string) (*Side, error) {
	side := Side(s)
	switch side {
	case SideBuy:
	case SideSell:
	default:
		return nil, fmt.Errorf("invalid side: %v", s)
	}
	return &side, nil
}

func (s Side) Opposite() Side {
	if s == SideBuy {
		return SideSell
	}
	return SideBuy
}

func (s Side) String() string {
	return string(s)
}

// Order Type
type OrderType string

func (t OrderType) String() string {
	return string(t)
}

func (t OrderType) Int() int64 {
	switch t {
	case "limit":
		return int64(2)
	case "market":
		return int64(1)
	case "stop order":
		return int64(3)
	}
	return 0
}

// Used to indicate order status
type OrderStatus string

func NewOrderStatusFromString(s string) (*OrderStatus, error) {
	status := OrderStatus(s)
	switch status {
	case OrderStatusNew:
	case OrderStatusOpen:
	case OrderStatusCancelling:
	case OrderStatusCancelled:
	case OrderStatusFilled:
	default:
		return nil, fmt.Errorf("invalid status: %v", s)
	}
	return &status, nil
}

func (t OrderStatus) String() string {
	return string(t)
}

// Used to indicate the type of bill
type BillType string

// Used to indicate the reason for a fill completion
type DoneReason string

type TransactionStatus string

const (
	OrderTypeLimit  = OrderType("limit")
	OrderTypeMarket = OrderType("market")

	SideBuy  = Side("buy")
	SideSell = Side("sell")

	// Initial state
	OrderStatusNew = OrderStatus("new")
	// Already joined orderBook
	OrderStatusOpen = OrderStatus("open")
	// Intermediate status, request cancellation
	OrderStatusCancelling = OrderStatus("cancelling")
	// The order has been canceled, and some orders are also canceled
	OrderStatusCancelled = OrderStatus("cancelled")
	// Completed order
	OrderStatusFilled = OrderStatus("filled")

	BillTypeTrade = BillType("trade")

	DoneReasonFilled    = DoneReason("filled")
	DoneReasonCancelled = DoneReason("cancelled")

	TransactionStatusPending   = TransactionStatus("pending")
	TransactionStatusCompleted = TransactionStatus("completed")
)

type User struct {
	Id           int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserId       int64
	Email        string
	PasswordHash string
}

type Account struct {
	Id        int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
	UserId    int64           `gorm:"column:user_id;unique_index:idx_uid_currency"`
	Currency  string          `gorm:"column:currency;unique_index:idx_uid_currency"`
	Hold      decimal.Decimal `gorm:"column:hold" sql:"type:decimal(32,16);"`
	Available decimal.Decimal `gorm:"column:available" sql:"type:decimal(32,16);"`
}

type Bill struct {
	Id        int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
	UserId    int64
	Currency  string
	Available decimal.Decimal `sql:"type:decimal(32,16);"`
	Hold      decimal.Decimal `sql:"type:decimal(32,16);"`
	Type      BillType
	Settled   bool
	Notes     string
}

type Product struct {
	Id             string `gorm:"column:id;primary_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	BaseCurrency   string
	QuoteCurrency  string
	BaseMinSize    decimal.Decimal `sql:"type:decimal(32,16);"`
	BaseMaxSize    decimal.Decimal `sql:"type:decimal(32,16);"`
	QuoteMinSize   decimal.Decimal `sql:"type:decimal(32,16);"`
	QuoteMaxSize   decimal.Decimal `sql:"type:decimal(32,16);"`
	BaseScale      int32
	QuoteScale     int32
	QuoteIncrement float64
}

/*type Order struct {
	Id             int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ProductId      string
	UserId         int64
	ClientOid      string
	Size           decimal.Decimal `sql:"type:decimal(32,16);"`
	Funds          decimal.Decimal `sql:"type:decimal(32,16);"`
	FilledSize     decimal.Decimal `sql:"type:decimal(32,16);"`
	ExecutedValue  decimal.Decimal `sql:"type:decimal(32,16);"`
	Price          decimal.Decimal `sql:"type:decimal(32,16);"`
	FillFees       decimal.Decimal `sql:"type:decimal(32,16);"`
	Type           OrderType
	Side           Side
	TimeInForce    string
	Status         OrderStatus
	Settled        bool
	ExpiresIn      int64
	BackendOrderId string
	Art            string
}*/
type GFill struct {
	Id         int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	TradeId    int64
	OrderId    int64 `gorm:"unique_index:o_m"`
	MessageSeq int64 `gorm:"unique_index:o_m"`
	ProductId  string
	Size       decimal.Decimal `sql:"type:decimal(32,16);"`
	Price      decimal.Decimal `sql:"type:decimal(32,16);"`
	Funds      decimal.Decimal `sql:"type:decimal(32,16);"`
	Fee        decimal.Decimal `sql:"type:decimal(32,16);"`
	Liquidity  string
	Settled    bool
	Side       Side
	Done       bool
	DoneReason DoneReason
	LogOffset  int64
	LogSeq     int64
	ExpiresIn  int64
	//	ClientOid  string
	Art         int64
	CancelledAt string
	ExecutedAt  string
}

type Fill struct {
	Id          int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TradeId     int64
	OrderId     int64 `gorm:"unique_index:o_m"`
	MessageSeq  int64 `gorm:"unique_index:o_m"`
	ProductId   string
	Size        decimal.Decimal `sql:"type:decimal(32,16);"`
	Price       decimal.Decimal `sql:"type:decimal(32,16);"`
	Funds       decimal.Decimal `sql:"type:decimal(32,16);"`
	Fee         decimal.Decimal `sql:"type:decimal(32,16);"`
	Liquidity   string
	Settled     bool
	Side        Side
	Done        bool
	DoneReason  DoneReason
	LogOffset   int64
	LogSeq      int64
	ClientOid   string
	ExpiresIn   int64
	Art         int64
	CancelledAt string
	ExecutedAt  string
}

type Trade struct {
	Id              int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ProductId       string
	TakerOrderId    int64
	MakerOrderId    int64
	Price           decimal.Decimal `sql:"type:decimal(32,16);"`
	Size            decimal.Decimal `sql:"type:decimal(32,16);"`
	Side            Side
	Time            time.Time
	LogOffset       int64
	LogSeq          int64
	TakerArt        int64
	MakerArt        int64
	TakerExecutedAt string
	MakerExecutedAt string
}

type Config struct {
	Id        int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Key       string
	Value     string
}

type Transaction struct {
	Id          int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UserId      int64
	Currency    string
	BlockNum    int
	ConfirmNum  int
	Status      TransactionStatus
	FromAddress string
	ToAddress   string
	Note        string
	TxId        string
}

type Expiry struct {
	OrderId   int64
	Timer     int64
	LogOffset int64
}

type PlaceOrderRequest struct {
	ClientOid      string  `json:"client_oid"`
	ProductId      string  `json:"productId"`
	UserId         int64   `json:"userId"`
	Size           float64 `json:"size"`
	Funds          float64 `json:"funds"`
	Price          float64 `json:"price"`
	Side           string  `json:"side"`
	Type           string  `json:"type"`        // [optional] limit or market (default is limit)
	TimeInForce    string  `json:"timeInForce"` // [optional] GTC, GTT, IOC, or FOK (default is GTC)
	ExpiresIn      int64   `json:"expiresIn"`   // [optional] set expiresIn except marker-order
	BackendOrderId string  `json:"backendOrderId"`
	Art            int64   `json:"art_name"`
}

type EstimateValue struct {
	Price    float64
	Quantity float64
}

type Order struct {
	Id             int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ProductId      string
	UserId         int64 `gorm:"column:user"`
	ClientOid      string
	Size           decimal.Decimal `gorm:"column:artBits" sql:"type:decimal(32,16);"`
	Funds          decimal.Decimal `gorm:"column:totalAmount" sql:"type:decimal(32,16);"`
	FilledSize     decimal.Decimal `gorm:"column:filledArtBits" sql:"type:decimal(32,16);"`
	ExecutedValue  decimal.Decimal `gorm:"column:filledAmount" sql:"type:decimal(32,16);"`
	Price          decimal.Decimal `sql:"type:decimal(32,16);"`
	FillFees       decimal.Decimal `gorm:"column:commission" sql:"type:decimal(32,16);"`
	Type           int64           `gorm:"column:orderType"`
	Side           Side
	Status         OrderStatus
	ExpiresIn      int64
	BackendOrderId string     `gorm:"column:orderId"`
	Art            int64      `gorm:"column:art"`
	CancelledAt    *time.Time `gorm:"column:cancelledAt;default:null"`
	ExecutedAt     *time.Time `gorm:"column:executedAt;default:null"`
	DeletedAt      *time.Time `gorm:"column:deletedAt;default:null"`
	UserRole       int64      `gorm:"column:userRole"`
	Settled        bool
}

type Tabler interface {
	TableName() string
}

func (Order) TableName() string {
	return "OrderBooks"
}

type OrderBookResponse struct {
	Ask      []map[string]decimal.Decimal `json:"ask"`
	Bid      []map[string]decimal.Decimal `json:"bid"`
	UsdSpace decimal.Decimal              `json:"usd_spread"`
}
