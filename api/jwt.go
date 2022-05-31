package api

import (
	"github.com/bCoder778/log"
	"github.com/dgrijalva/jwt-go"
)

func validateToken(tokenString string, secretKey string) bool {
	_, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		log.Infof("validate token failed! %s", err)
		return false
	}
	return true
}
