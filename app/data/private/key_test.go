package private_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/keratin/authn-server/app/data/private"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKey(t *testing.T) {
	rnd := rand.New(rand.NewSource(1234))
	rsaKey, err := rsa.GenerateKey(rnd, 256)
	require.NoError(t, err)
	key, err := private.NewKey(rsaKey)
	require.NoError(t, err)
	assert.Equal(t, rsaKey, key.PrivateKey)
	assert.Len(t, key.JWK.KeyID, 43)
}

func TestGenerate(t *testing.T) {
	bitLen := 512
	key, err := private.GenerateKey(bitLen)
	require.NoError(t, err)
	assert.Equal(t, key.PrivateKey.N.BitLen(), bitLen, "generated key should have requested bit length")
}

func TestKeyID(t *testing.T) {
	expected := "zUqHmeEoXm1691utTlGRdXjHhG-QEgEY12oQgYK3K0s"
	serialized := []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAteQCTn7AGpoFG+IDZ3UNwHsG8KYXIw7JRp/vlN8/Gqj9FLZj\neq0NBUe9u1qLQOy0PLKAV9BkmT1iKbCJdSbhWz39apWFw95thYRGX5xapF6VI4jr\noUsdINNbOUdRkc9mDsqDZo0d8JGsWv1hA+2SGuI6P4s7Lf/cmzYFh++Jy01qBwyd\n5mHSayDD2oApcGlkOOjfUbb2DZveceKpD+0rq4abICzI1GfuC/Gnedw7hE4XChJ4\nHKuBEcHC9leBQmZ40PF35YBHaja0Uy8QtNalILRPFAC2jEXJkyez9Z27clVReD/n\nPcfZ/9bs7W1+VfZrVSIIHda55gYJRCTr8MMsLQIDAQABAoIBAQCHj6PYdMcgDGJ6\nYXxAAxF4vzhw6pib3E1OgazBu5EAgan9YeHKcGcf5FQX6meWv9Ok2TSmPf575y/d\n+mC4G34hzpWsdjv3uzLNK8R3RcSYdJWaolVbJOxUprF6gxjcH0LlCzHboJkLzsYy\nGl3P26PkvW7EJTS6F9OHKj/9DB4akhoX6af2C1ivL/TPg+GDUFMNNeEvXk/WhAdC\nm/RPLmTiFi9h0AsS40NW3zX8Dft3mBEWaNsrcQH5CJz23ozpbq6pA4ZqLXUlAtrI\nPXx5FCM5q+VJPHE1cnmthZHVXsOo7uwORm7U2ry2PUzBJ4kQx3LfybRNcIfU8K5y\nejoXrjwBAoGBANN6PEWbtyLAsXLVfSFI/Jk36+Sm6uJh0vnDHP6KlhHY3yzP5ruo\nSh7F2vRErYk+uK7k43oyzG0PyJCnket1g3c8oZAn11pm7GtyEAJOGcGrKoiCjR4x\n5gtDbZgGxr07CN/RyE+ZbNMoutWXRrPUfLxNPRRLdWWfKlwdfH/PDDttAoGBANwv\nKnOsUu9CZ5Sc0Eq4nev+h8t9dAIehnfB2qhNmhUv6PAbJu9odioKLdyN0kP33rIj\nPp+MxCbtmmeU7Bk1CqEEe5lC1A3EivY9w0+b7A64hmfiuP8XVJ2gojvTajOoGJ0w\nC6OolQx+4TTiRzDakpdzFXbbDiJ0+k8blcLE0XvBAoGACsB4OAHGudmaLAB2sC6J\nyTByqdlir8fRdilZXAenwZiJIDohvQC9Y/sjOrATMpshwKKafif/BLx8sf4TCSmc\nWX+Xp0CfTlVVR9Ewxy05WgNd0jrw+cwHqiLve388s3pA5UBBMurWAZZciWd7jMEM\n5nX22QVNHrGM8cn9/nGEabECgYB8LzHvSbsA7OAExqkH67ZOGyG12IzsgRDwTGqp\n0BLebkYf3gCIuM8kiNcy9N4prYxxxkUUsc0T86DJWQoMcYkMJb4cQ7/cAAUsOsuE\ng/mQl+xefVY/sYXs3WODAIt+lQlE5os6A+QExy73p8PlPvG875Ckl4oSTw26PmGq\nF13bQQKBgA4wO1Xq5TaRWa8R5jb5f1Y/8JX5GybWaD1H2wB0lX/W/E064Z5hIgoM\n+DwqzJKNkhEEdtHXMJS10qV/qcGmOHOE1FUnWjQ/20s6GpImb45H84aJwvKmvPf8\nmn0NHZRqGdMecEfREyT2oYEQ+pfdJ4t6LTJGiHxYYt7WzpzOZ1cU\n-----END RSA PRIVATE KEY-----\n")
	block, _ := pem.Decode(serialized)
	rsaKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	key, err := private.NewKey(rsaKey)
	require.NoError(t, err)
	assert.Equal(t, expected, key.JWK.KeyID)
}
