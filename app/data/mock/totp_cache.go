package mock

import (
	"fmt"
)

type TOTP struct {
	store     map[int][]byte
	errorOnID int
}

func NewTOTPCache(errorOnID int) *TOTP {
	return &TOTP{
		errorOnID: errorOnID,
		store:     make(map[int][]byte),
	}
}

func (m TOTP) CacheTOTPSecret(accountID int, secret []byte) error {
	if accountID == m.errorOnID {
		return fmt.Errorf("error forced by ID: %d", accountID)
	}
	m.store[accountID] = secret
	return nil
}

func (m TOTP) LoadTOTPSecret(accountID int) ([]byte, error) {
	if accountID == m.errorOnID {
		return nil, fmt.Errorf("error forced by ID: %d", accountID)
	}
	r, ok := m.store[accountID]
	if !ok {
		return nil, nil
	}
	return r, nil
}

func (m TOTP) RemoveTOTPSecret(accountID int) error {
	if accountID == m.errorOnID {
		return fmt.Errorf("error forced by ID: %d", accountID)
	}
	delete(m.store, accountID)
	return nil
}
