package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
)

func TestFieldErrors(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		fe := services.FieldErrors{{"username", "blank"}, {"password", "blank"}}
		assert.Equal(t, "username: blank, password: blank", fe.Error())
	})
}

func TestUsernameValidator(t *testing.T) {
	t.Run("email usernames", func(t *testing.T) {
		cfg := &app.Config{UsernameIsEmail: true}

		t.Run("good emails", func(t *testing.T) {
			emails := []string{
				"foo@bar.tld",
				"foo@bar.baz.tld",
				"foo.bar@baz.tld",
			}

			for _, email := range emails {
				err := services.UsernameValidator(cfg, email)
				assert.Nil(t, err, email)
			}
		})

		t.Run("bad emails", func(t *testing.T) {
			emails := []string{
				"foo@bar",
				"foo@bar..tld",
				"foo@.bar.tld",
			}

			for _, email := range emails {
				err := services.UsernameValidator(cfg, email)
				assert.NotNil(t, err, email)
				assert.Error(t, err)
			}
		})

	})
}
