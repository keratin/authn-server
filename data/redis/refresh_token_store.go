package redis

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/keratin/authn/data"
)

type RefreshTokenStore struct {
	Client *redis.Client
	TTL    time.Duration
}

// Redis key for token => account_id lookup
func keyForToken(t []byte) string {
	str := fmt.Sprintf("s:t.%s", t)
	return str
}

// Redis key for account_id => tokens lookup
func keyForAccount(id int) string {
	str := fmt.Sprintf("s:a.%d", id)
	return str
}

func (s *RefreshTokenStore) Find(hexToken data.RefreshToken) (int, error) {
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

func (s *RefreshTokenStore) Touch(hexToken data.RefreshToken, account_id *int) error {
	binToken, err := hex.DecodeString(string(hexToken))
	if err != nil {
		return err
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.Expire(keyForToken(binToken), s.TTL)
		pipe.Expire(keyForAccount(*account_id), s.TTL)
		return nil
	})
	return err
}

func (s *RefreshTokenStore) FindAll(account_id int) ([]data.RefreshToken, error) {
	bins, err := s.Client.SMembers(keyForAccount(account_id)).Result()
	if err != nil {
		return nil, err
	}

	tokens := make([]data.RefreshToken, 0)
	for _, t := range bins {
		tokens = append(tokens, data.RefreshToken(hex.EncodeToString([]byte(t))))
	}

	return tokens, nil
}

func (s *RefreshTokenStore) Create(account_id int) (data.RefreshToken, error) {
	binToken := generateToken()

	_, err := s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		// persist the token
		pipe.Set(keyForToken(binToken), account_id, s.TTL)

		// maintain a list of tokens per account_id
		pipe.SAdd(keyForAccount(account_id), binToken)
		pipe.Expire(keyForAccount(account_id), s.TTL)

		return nil
	})
	if err != nil {
		return "", err
	}

	return data.RefreshToken(hex.EncodeToString(binToken)), nil
}

func (s *RefreshTokenStore) Revoke(hexToken data.RefreshToken) error {
	account_id, err := s.Find(hexToken)
	if err != nil {
		return err
	}

	_, err = s.Client.Pipelined(func(pipe redis.Pipeliner) error {
		binToken, err := hex.DecodeString(string(hexToken))
		if err != nil {
			return err
		}

		pipe.Del(keyForToken(binToken))
		pipe.SRem(keyForAccount(account_id), binToken)

		return nil
	})
	return err
}

// 128 bits of randomness is more than a UUID v4
// cf: https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_.28random.29
func generateToken() []byte {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}
	return token
}
