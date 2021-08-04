package models

type Store interface {
	BeginTx() (Store, error)
	Rollback() error
	CommitTx() error

	GetAccount(userId int64, currency string) (*Account, error)
	GetAccountForUpdate(userId int64, currency string) (*Account, error)
	UpdateAccount(account *Account) error

	AddBills(bills []*Bill) error

	AddFills(fills []*Fill) error
	GetLastFillByProductId(productId string) (*Fill, error)

	AddOrder(order *Order) error
	UpdateOrderStatus(orderId int64, oldStatus, newStatus OrderStatus) (bool, error)

	GetProductById(id string) (*Product, error)
	GetProducts() ([]*Product, error)
}
