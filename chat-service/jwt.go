package main

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Plan   string `json:"plan"`
	Age  int    `json:"age"`
	Gender string `json:"gender"`
	jwt.RegisteredClaims
}

type JWT struct {
	AccessSecret  string
}

func NewJWT(accessSecret string) *JWT {
	return &JWT{
		AccessSecret:  accessSecret,
	}
}


func (j *JWT) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {		
		return []byte(j.AccessSecret), nil
	})

	if err != nil {
		return nil, errors.New("invalid token: " + err.Error())
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}