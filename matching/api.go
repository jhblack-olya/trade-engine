package matching

import "gitlab.com/gae4/trade-engine/models"

type OrderReader interface {
	SetOffset(int64) error
	FetchOrder() (int64, *models.Order, error)
}

type LogStore interface {
	Store(logs []interface{}) error
}

type LogReader interface {
	GetProductId() string
	RegisterObserver(observer LogObserver)
	Run(seq, offset int64)
}

// Match log reader observer
type LogObserver interface {
	OnOpenLog(log *OpenLog, offset int64)
	OnMatchLog(log *MatchLog, offset int64)
	OnDoneLog(log *DoneLog, offset int64)
}

type SnapshotStore interface {
	Store(snapshot *Snapshot) error
	GetLatest() (*Snapshot, error)
}