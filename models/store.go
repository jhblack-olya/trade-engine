package models

type Store interface {
	BeginTx() (Store, error)
	Rollback() error
	CommitTx() error

	GetAccount(userId int64, currency string) (*Account, error)
	GetAccountForUpdate(userId int64, currency string) (*Account, error)
	UpdateAccount(account *Account) error

	AddBills(bills []*Bill) error

	GetProductById(id string) (*Product, error)
	GetProducts() ([]*Product, error)

	AddOrder(order *Order) error

	GetTicksByProductId(productId string, granularity int64, limit int) ([]*Tick, error)
}
