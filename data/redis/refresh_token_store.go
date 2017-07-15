package redis

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/keratin/authn-server/models"
)

type RefreshTokenStore struct {
	Client *redis.Client
	TTL    time.Duration
}

// Redis key for token => accountId lookup
func keyForToken(t []byte) string {
	str := fmt.Sprintf("s:t.%s", t)
	return str
}

// Redis key for accountId => tokens lookup
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

func (s *RefreshTokenStore) Touch(hexToken models.RefreshToken, accountId int) error {
	binToken, err := hex.DecodeString(string(hexToken))
	if err != nil {
		return err
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.Expire(keyForToken(binToken), s.TTL)
		pipe.Expire(keyForAccount(accountId), s.TTL)
		return nil
	})
	return err
}

func (s *RefreshTokenStore) FindAll(accountId int) ([]models.RefreshToken, error) {
	bins, err := s.Client.SMembers(keyForAccount(accountId)).Result()
	if err != nil {
		return nil, err
	}

	tokens := make([]models.RefreshToken, 0)
	for _, t := range bins {
		tokens = append(tokens, models.RefreshToken(hex.EncodeToString([]byte(t))))
	}

	return tokens, nil
}

func (s *RefreshTokenStore) Create(accountId int) (models.RefreshToken, error) {
	binToken, err := generateToken()
	if err != nil {
		return "", err
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		// persist the token
		pipe.Set(keyForToken(binToken), accountId, s.TTL)

		// maintain a list of tokens per accountId
		pipe.SAdd(keyForAccount(accountId), binToken)
		pipe.Expire(keyForAccount(accountId), s.TTL)

		return nil
	})
	if err != nil {
		return "", err
	}

	return models.RefreshToken(hex.EncodeToString(binToken)), nil
}

func (s *RefreshTokenStore) Revoke(hexToken models.RefreshToken) error {
	accountId, err := s.Find(hexToken)
	if err != nil {
		return err
	}
	if accountId == 0 {
		return nil
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		binToken, err := hex.DecodeString(string(hexToken))
		if err != nil {
			return err
		}

		pipe.Del(keyForToken(binToken))
		pipe.SRem(keyForAccount(accountId), binToken)

		return nil
	})
	return err
}

// 128 bits of randomness is more than a UUID v4
// cf: https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_.28random.29
func generateToken() ([]byte, error) {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		return []byte{}, nil
	}
	return token, nil
}
