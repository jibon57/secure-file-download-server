package main

import (
	"fmt"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"time"
)

func main() {
	path := "asset/file/test.pdf"
	apiKey := "API3Q37hsqU"
	apiSecret := "zW8EyC44kI4kLKeAcSOdYsw6BOy9mu"

	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(apiSecret)}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		panic(err)
	}

	cl := jwt.Claims{
		Issuer:    apiKey,
		NotBefore: jwt.NewNumericDate(time.Now()),
		Expiry:    jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		Subject:   path,
	}

	token, err := jwt.Signed(sig).Claims(cl).CompactSerialize()
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
