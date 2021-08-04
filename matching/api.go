package matching

import "gitlab.com/gae4/trade-engine/models"

type orderBookSnapshot struct {
	ProductId     string
	Orders        []BookOrder
	TradeSeq      int64
	LogSeq        int64
	OrderIdWindow Window
}

type OrderReader interface {
	SetOffset(int64) error
	FetchOrder() (int64, *models.Order, error)
}

type LogStore interface {
	Store(logs []interface{}) error
}

type SnapshotStore interface {
	Store(snapshot *Snapshot) error
	GetLatest() (*Snapshot, error)
}

type LogReader interface {
	GetProductId() string
	RegisterObserver(observer LogObserver)
	Run(seq, offset int64)
}

type LogObserver interface {
	OnOpenLog(log *OpenLog, offset int64)
	OnMatchLog(log *MatchLog, offset int64)
	OnDoneLog(log *DoneLog, offset int64)
}
