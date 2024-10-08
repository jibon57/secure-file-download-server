package main

import (
	"fmt"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
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
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		Expiry:    jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
		Subject:   path,
	}

	token, err := jwt.Signed(sig).Claims(cl).Serialize()
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
