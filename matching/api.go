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

type LogReader interface {
	// Get the current productId
	GetProductId() string

	// Register a log observer
	RegisterObserver(observer LogObserver)

	// Start to read the log, the read log will be callback to the observer
	Run(seq, offset int64)
}

// Match log reader observer
type LogObserver interface {
	// Callback when OpenLog is read
	OnOpenLog(log *OpenLog, offset int64)

	// Called when MatchLog is read
	OnMatchLog(log *MatchLog, offset int64)

	// Callback when DoneLog is read
	OnDoneLog(log *DoneLog, offset int64)
}

type SnapshotStore interface {
	Store(snapshot *Snapshot) error
	GetLatest() (*Snapshot, error)
}
