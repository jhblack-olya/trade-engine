package mysql

import (
	"time"

	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) AddAccount(account *models.Account) error {
	account.CreatedAt = time.Now()
	return s.db.Create(account).Error
}
