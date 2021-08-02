package mysql

import (
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) GetAccount(userId int64, currency string) (*models.Account, error) {
	var account models.Account
	err := s.db.Raw("SELECT * FROM g_account WHERE user_id=? AND currency=?",
		userId, currency).Scan(&account).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &account, err
}

func (s *Store) GetAccountForUpdate(userId int64, currency string) (*models.Account, error) {
	var account models.Account
	err := s.db.Raw("SELECT * FROM g_account WHERE user_id=? AND currency=? FOR UPDATE",
		userId, currency).Scan(&account).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &account, err
}

func (s *Store) UpdateAccount(account *models.Account) error {
	account.UpdatedAt = time.Now()
	return s.db.Save(account).Error
}

func (s *Store) AddAccount(account *models.Account) error {
	account.CreatedAt = time.Now()
	return s.db.Create(account).Error
}
