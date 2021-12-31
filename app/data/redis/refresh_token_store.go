package redis

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/lib"
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
	str, err := s.Client.Get(context.TODO(), keyForToken(binToken)).Result()
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

	_, err = s.Client.Pipelined(context.TODO(), func(pipe redis.Pipeliner) error {
		pipe.Expire(context.TODO(), keyForToken(binToken), s.TTL)
		pipe.Expire(context.TODO(), keyForAccount(accountID), s.TTL)
		return nil
	})
	return err
}

func (s *RefreshTokenStore) FindAll(accountID int) ([]models.RefreshToken, error) {
	bins, err := s.Client.SMembers(context.TODO(), keyForAccount(accountID)).Result()
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

	_, err = s.Client.Pipelined(context.TODO(), func(pipe redis.Pipeliner) error {
		// persist the token
		pipe.Set(context.TODO(), keyForToken(binToken), accountID, s.TTL)

		// maintain a list of tokens per accountID
		pipe.SAdd(context.TODO(), keyForAccount(accountID), binToken)
		pipe.Expire(context.TODO(), keyForAccount(accountID), s.TTL)

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

	_, err = s.Client.Pipelined(context.TODO(), func(pipe redis.Pipeliner) error {
		binToken, err := hex.DecodeString(string(hexToken))
		if err != nil {
			return err
		}

		pipe.Del(context.TODO(), keyForToken(binToken))
		pipe.SRem(context.TODO(), keyForAccount(accountID), binToken)

		return nil
	})
	return err
}
