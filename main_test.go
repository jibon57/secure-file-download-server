package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/jibon57/secure-file-download-server/internal"
)

func Test_Main(t *testing.T) {
	test_readYaml(t)
	file := "test.txt"

	_, err := os.Lstat(fmt.Sprintf("%s/%s", internal.AppCnf.Path, file))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = internal.CreateOrUpdateDirs()
			if err != nil {
				t.Errorf("can't create directory %s", err.Error())
			}
			emptyFile, err := os.Create(fmt.Sprintf("%s/%s", internal.AppCnf.Path, file))
			if err != nil {
				t.Errorf("can't create test file %s", err.Error())
				return
			}
			_ = emptyFile.Close()
		}
	}

	token, err := genToken(file)
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name   string
		method string
		route  string
		body   string
	}{
		{
			name:   "download",
			route:  "/download/" + token,
			method: http.MethodGet,
		},
		{
			name:   "delete",
			route:  "/delete",
			method: http.MethodPost,
			body:   `{"file_path": "test.txt"}`,
		},
	}

	r := internal.Router(Version)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.route, strings.NewReader(tt.body))
			if tt.method == http.MethodPost {
				req.Header.Set("content-type", "application/json")
				req.Header.Set("API-KEY", internal.AppCnf.ApiKey)
				req.Header.Set("API-SECRET", internal.AppCnf.ApiSecret)
			}

			res, err := r.Test(req)
			if err != nil {
				t.Error(err)
			}

			if res.StatusCode != 200 {
				t.Errorf("Route: %s, Error code: %d", tt.route, res.StatusCode)
			}
		})
	}
}

func test_readYaml(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "readYaml",
			args: args{
				filename: "config_sample.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := readYaml(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("readYaml() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func genToken(file string) (string, error) {
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(internal.AppCnf.ApiSecret)}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}
	cl := jwt.Claims{
		Issuer:    internal.AppCnf.ApiKey,
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		Expiry:    jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 30)),
		Subject:   file,
	}

	token, err := jwt.Signed(sig).Claims(cl).Serialize()
	if err != nil {
		return "", err
	}

	return token, nil
}
