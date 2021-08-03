package models

type Store interface {
	BeginTx() (Store, error)
	Rollback() error
	CommitTx() error

	AddAccount(account *Account) error
	GetAccount(userId int64, currency string) (*Account, error)
	GetAccountForUpdate(userId int64, currency string) (*Account, error)
	UpdateAccount(account *Account) error

	AddOrder(order *Order) error

	AddBills(bills []*Bill) error
	GetUnsettledBillsByUserId(userId int64, currency string) ([]*Bill, error)
	GetUnsettledBills() ([]*Bill, error)
	UpdateBill(bill *Bill) error

	GetProductById(id string) (*Product, error)
	GetProducts() ([]*Product, error)
}
