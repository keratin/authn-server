package services_test

import (
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFieldErrors(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		fe := services.FieldErrors{{"username", "blank"}, {"password", "blank"}}
		assert.Equal(t, "username: blank, password: blank", fe.Error())
	})
}
