package tokens

import jwt "github.com/dgrijalva/jwt-go"

func Sign(claims jwt.Claims, method jwt.SigningMethod, key interface{}) (string, error) {
	return jwt.NewWithClaims(method, claims).SignedString(key)
}
