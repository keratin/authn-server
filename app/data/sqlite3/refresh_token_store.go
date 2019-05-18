package sqlite3

import (
	"database/sql"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/app/models"
)

type RefreshTokenStore struct {
	sqlx.Ext
	TTL time.Duration
}

func (s *RefreshTokenStore) Clean(reporter ops.ErrorReporter) {
	go func() {
		for range time.Tick(time.Minute + time.Duration(rand.Intn(5))*time.Second) {
			_, err := s.Exec("DELETE FROM refresh_tokens WHERE expires_at < ?", time.Now())
			if err != nil {
				reporter.ReportError(errors.Wrap(err, "RefreshTokenStore Clean"))
			}
			time.Sleep(time.Minute)
		}
	}()
}

func (s *RefreshTokenStore) Create(accountID int) (models.RefreshToken, error) {
	binToken, err := lib.GenerateToken()
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(binToken)

	_, err = s.Exec(
		"INSERT INTO refresh_tokens (account_id, token, expires_at) VALUES (?, ?, ?)",
		accountID,
		token,
		time.Now().Add(s.TTL),
	)
	if err != nil {
		return "", err
	}
	return models.RefreshToken(token), nil
}

func (s *RefreshTokenStore) Find(token models.RefreshToken) (int, error) {
	var accountID int
	err := s.QueryRowx(
		"SELECT account_id FROM refresh_tokens WHERE token = ? AND expires_at > ?",
		token,
		time.Now(),
	).Scan(&accountID)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return accountID, nil
}

func (s *RefreshTokenStore) Touch(token models.RefreshToken, accountID int) error {
	_, err := s.Exec(
		"UPDATE refresh_tokens SET expires_at = ? WHERE token = ? AND expires_at > ?",
		time.Now().Add(s.TTL),
		token,
		time.Now(),
	)
	return err
}

func (s *RefreshTokenStore) FindAll(accountID int) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	rows, err := s.Query(
		"SELECT token FROM refresh_tokens WHERE account_id = ? AND expires_at > ?",
		accountID,
		time.Now(),
	)
	if err != nil {
		return tokens, err
	}
	defer rows.Close()
	for rows.Next() {
		var token string
		err := rows.Scan(&token)
		if err != nil {
			return []models.RefreshToken{}, err
		}
		tokens = append(tokens, models.RefreshToken(token))
	}

	return tokens, nil
}

func (s *RefreshTokenStore) Revoke(token models.RefreshToken) error {
	_, err := s.Exec("DELETE FROM refresh_tokens WHERE token = ?", token)
	return err
}
