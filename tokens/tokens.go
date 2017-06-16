package tokens

import jwt "github.com/dgrijalva/jwt-go"

type Claimser interface {
	claims() *jwt.MapClaims
}

func Sign(tkn Claimser, method jwt.SigningMethod, key interface{}) (string, error) {
	return jwt.NewWithClaims(method, tkn.claims()).SignedString(key)
}
