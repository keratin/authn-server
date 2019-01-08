package redis

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/app/models"
)

type RefreshTokenStore struct {
	*redis.Client
	TTL time.Duration
}

// Redis key for token => accountID lookup
func keyForToken(t []byte) string {
	str := fmt.Sprintf("s:t.%s", t)
	return str
}

// Redis key for accountID => tokens lookup
func keyForAccount(id int) string {
	str := fmt.Sprintf("s:a.%d", id)
	return str
}

func (s *RefreshTokenStore) Find(hexToken models.RefreshToken) (int, error) {
	binToken, err := hex.DecodeString(string(hexToken))
	if err != nil {
		return 0, err
	}
	str, err := s.Client.Get(keyForToken(binToken)).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (s *RefreshTokenStore) Touch(hexToken models.RefreshToken, accountID int) error {
	binToken, err := hex.DecodeString(string(hexToken))
	if err != nil {
		return err
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.Expire(keyForToken(binToken), s.TTL)
		pipe.Expire(keyForAccount(accountID), s.TTL)
		return nil
	})
	return err
}

func (s *RefreshTokenStore) FindAll(accountID int) ([]models.RefreshToken, error) {
	bins, err := s.Client.SMembers(keyForAccount(accountID)).Result()
	if err != nil {
		return nil, err
	}

	tokens := make([]models.RefreshToken, 0)
	for _, t := range bins {
		tokens = append(tokens, models.RefreshToken(hex.EncodeToString([]byte(t))))
	}

	return tokens, nil
}

func (s *RefreshTokenStore) Create(accountID int) (models.RefreshToken, error) {
	binToken, err := lib.GenerateToken()
	if err != nil {
		return "", err
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		// persist the token
		pipe.Set(keyForToken(binToken), accountID, s.TTL)

		// maintain a list of tokens per accountID
		pipe.SAdd(keyForAccount(accountID), binToken)
		pipe.Expire(keyForAccount(accountID), s.TTL)

		return nil
	})
	if err != nil {
		return "", err
	}

	return models.RefreshToken(hex.EncodeToString(binToken)), nil
}

func (s *RefreshTokenStore) Revoke(hexToken models.RefreshToken) error {
	accountID, err := s.Find(hexToken)
	if err != nil {
		return err
	}
	if accountID == 0 {
		return nil
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		binToken, err := hex.DecodeString(string(hexToken))
		if err != nil {
			return err
		}

		pipe.Del(keyForToken(binToken))
		pipe.SRem(keyForAccount(accountID), binToken)

		return nil
	})
	return err
}
