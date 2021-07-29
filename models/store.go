package models

type Store interface {
	BeginTx() (Store, error)
	Rollback() error
	CommitTx() error

	GetProducts() ([]*Product, error)
	GetProductById(string) (*Product, error)

	GetAccount(userId int64, currency string) (*Account, error)
	GetAccountForUpdate(userId int64, currency string) (*Account, error)
	AddAccount(account *Account) error
	UpdateAccount(account *Account) error

	AddBills(bills []*Bill) error

	AddOrder(order *Order) error
}
