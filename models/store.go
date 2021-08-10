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
	GetOrderById(orderId int64) (*Order, error)
	GetOrderByIdForUpdate(orderId int64) (*Order, error)
	UpdateOrder(order *Order) error

	AddBills(bills []*Bill) error
	GetUnsettledBillsByUserId(userId int64, currency string) ([]*Bill, error)
	GetUnsettledBills() ([]*Bill, error)
	UpdateBill(bill *Bill) error

	GetUnsettledFillsByOrderId(orderId int64) ([]*Fill, error)
	UpdateFill(fill *Fill) error
	GetUnsettledFills(count int32) ([]*Fill, error)

	AddFills(fills []*Fill) error
	GetLastFillByProductId(productId string) (*Fill, error)

	UpdateOrderStatus(orderId int64, oldStatus, newStatus OrderStatus) (bool, error)

	GetProductById(id string) (*Product, error)
	GetProducts() ([]*Product, error)

	GetLastTradeByProductId(productId string) (*Trade, error)
	AddTrades(trades []*Trade) error
}
