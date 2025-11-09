package internal

import (
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Router(version string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "File download server version: " + version,
	})
	if AppCnf.Debug {
		app.Use(logger.New())
	}
	app.Use(recover.New())

	// format: http://ip:port/download/token
	// make sure token is urlencoded
	app.Get("/download/:token", HandleDownloadFile)
	app.Get("/serveFile/:token", HandleServeFile)
	app.Post("/delete", HandleDeleteFile)

	return app
}

func verifyToken(token string) (*jwt.Claims, error) {
	tok, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.HS256})
	if err != nil {
		return nil, err
	}

	out := jwt.Claims{}
	if err = tok.Claims([]byte(AppCnf.ApiSecret), &out); err != nil {
		return nil, err
	}

	if err = out.Validate(jwt.Expected{Issuer: AppCnf.ApiKey, Time: time.Now().UTC()}); err != nil {
		return nil, err
	}

	return &out, nil
}

func sendResponse(c *fiber.Ctx, statusCode int, status bool, msg string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status": status,
		"msg":    msg,
	})
}
