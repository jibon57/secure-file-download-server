package main

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gopkg.in/square/go-jose.v2/jwt"
	"os"
	"strings"
	"time"
)

func Router() *fiber.App {
	app := fiber.New(fiber.Config{})
	if AppCnf.Debug {
		app.Use(logger.New())
	}
	app.Use(recover.New())

	// format: http://ip:port/download/token
	// make sure token is urlencoded
	app.Get("/download/:token", HandleDownloadFile)

	return app
}

func HandleDownloadFile(c *fiber.Ctx) error {
	token := c.Params("token")

	if len(token) == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": false,
			"msg":    "token require or invalid url",
		})
	}

	file, status, err := verifyToken(token)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"status": false,
			"msg":    err.Error(),
		})
	}

	c.Attachment(file)
	return c.SendFile(file, AppCnf.Compress)
}

func verifyToken(token string) (string, int, error) {
	tok, err := jwt.ParseSigned(token)
	if err != nil {
		return "", fiber.StatusUnauthorized, err
	}

	out := jwt.Claims{}
	if err = tok.Claims([]byte(AppCnf.ApiSecret), &out); err != nil {
		return "", fiber.StatusUnauthorized, err
	}

	if err = out.Validate(jwt.Expected{Issuer: AppCnf.ApiKey, Time: time.Now()}); err != nil {
		return "", fiber.StatusUnauthorized, err
	}

	file := fmt.Sprintf("%s/%s", AppCnf.Path, out.Subject)
	_, err = os.Lstat(file)

	if err != nil {
		ms := strings.SplitN(err.Error(), "/", -1)
		return "", fiber.StatusNotFound, errors.New(ms[len(ms)-1])
	}

	return file, fiber.StatusOK, nil
}
