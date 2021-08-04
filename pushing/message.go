package pushing

import "fmt"

type Level2Type string
type Channel string

func (t Channel) Format(productId string, userId int64) string {
	return fmt.Sprintf("%v:%v:%v", t, productId, userId)
}

func (t Channel) FormatWithUserId(userId int64) string {
	return fmt.Sprintf("%v:%v", t, userId)
}

const (
	Level2TypeSnapshot = Level2Type("snapshot")
	Level2TypeUpdate   = Level2Type("l2update")

	ChannelTicker = Channel("ticker")
	ChannelMatch  = Channel("match")
	ChannelLevel2 = Channel("level2")
	ChannelFunds  = Channel("funds")
	ChannelOrder  = Channel("order")
)

type Level2Change struct {
	Seq       int64
	ProductId string
	Side      string
	Price     string
	Size      string
}

type FundsMessage struct {
	Type      string `json:"type"`
	Sequence  int64  `json:"sequence"`
	UserId    string `json:"userId"`
	Currency  string `json:"currencyCode"`
	Available string `json:"available"`
	Hold      string `json:"hold"`
}

type OrderMessage struct {
	UserId        int64  `json:"userId"`
	Type          string `json:"type"`
	Sequence      int64  `json:"sequence"`
	Id            string `json:"id"`
	Price         string `json:"price"`
	Size          string `json:"size"`
	Funds         string `json:"funds"`
	ProductId     string `json:"productId"`
	Side          string `json:"side"`
	OrderType     string `json:"orderType"`
	CreatedAt     string `json:"createdAt"`
	FillFees      string `json:"fillFees"`
	FilledSize    string `json:"filledSize"`
	ExecutedValue string `json:"executedValue"`
	Status        string `json:"status"`
	Settled       bool   `json:"settled"`
}
