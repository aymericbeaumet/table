package jwt

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const tableTokenSubject = "table"

type TableToken struct {
	TableName string `json:"tbl"`
	jwt.StandardClaims
}

func NewTableToken(tableName string) *TableToken {
	return &TableToken{
		TableName: tableName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(3 * time.Minute).UTC().Unix(),
			Subject:   tableTokenSubject,
		},
	}
}

func ParseTableToken(tokenString string, key *ecdsa.PublicKey) (*TableToken, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&TableToken{},
		func(_ *jwt.Token) (interface{}, error) { return key, nil },
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token %#v", token)
	}

	claims, ok := token.Claims.(*TableToken)
	if !ok {
		return nil, fmt.Errorf("invalid claims type for token %#v", token)
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return nil, fmt.Errorf("expired token %#v", token)
	}

	if claims.Subject != tableTokenSubject {
		return nil, fmt.Errorf("unexpected subject for token %#v", token)
	}

	return claims, nil
}

func (c *TableToken) Sign(key *ecdsa.PrivateKey) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodES256, c).SignedString(key)
}
