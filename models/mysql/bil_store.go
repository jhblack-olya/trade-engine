package mysql

import (
	"fmt"
	"strings"

	"gitlab.com/gae4/trade-engine/models"
)

func (s *Store) AddBills(bills []*models.Bill) error {
	if len(bills) == 0 {
		return nil
	}
	var valueStrings []string
	for _, bill := range bills {
		valueString := fmt.Sprintf("(NOW(),%v, '%v', %v, %v, '%v', %v, '%v')",
			bill.UserId, bill.Currency, bill.Available, bill.Hold, bill.Type, bill.Settled, bill.Notes)
		valueStrings = append(valueStrings, valueString)
	}
	sql := fmt.Sprintf("INSERT INTO g_bill (created_at, user_id,currency,available,hold, type,settled,notes) VALUES %s", strings.Join(valueStrings, ","))
	return s.db.Exec(sql).Error
}
