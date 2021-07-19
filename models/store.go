package models

type Store interface {
	BeginTx() (Store, error)
	Rollback() error
	CommitTx() error

	GetProducts() ([]*Product, error)
}
