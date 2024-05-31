package main

import (
	"errors"
	"fmt"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"os"
	"strings"
	"time"
)

func Router() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "File download server version: " + Version,
	})
	if AppCnf.Debug {
		app.Use(logger.New())
	}
	app.Use(recover.New())

	// format: http://ip:port/download/token
	// make sure token is urlencoded
	app.Get("/download/:token", HandleDownloadFile)
	app.Post("/delete", HandleDeleteFile)

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

type DeleteFileReq struct {
	FilePath *string `json:"file_path,omitempty" xml:"file_path,omitempty" form:"file_path,omitempty"`
}

// HandleDeleteFile will require API-KEY & API-SECRET as header value
func HandleDeleteFile(c *fiber.Ctx) error {
	apiKey := c.Get("API-KEY")
	secret := c.Get("API-SECRET")

	if apiKey == "" || secret == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": false,
			"msg":    "Auth header information are missing",
		})
	}

	if apiKey != AppCnf.ApiKey || secret != AppCnf.ApiSecret {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": false,
			"msg":    "Auth header information didn't match",
		})
	}

	req := new(DeleteFileReq)
	err := c.BodyParser(req)
	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status": false,
			"msg":    err.Error(),
		})
	}

	if req.FilePath == nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status": false,
			"msg":    "file_path value require",
		})
	}

	file := fmt.Sprintf("%s/%s", AppCnf.Path, *req.FilePath)
	f, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, err.(*os.PathError)) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status": false,
				"msg":    *req.FilePath + " does not exist",
			})
		} else {
			return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
				"status": false,
				"msg":    err.Error(),
			})
		}
	}

	if f.IsDir() {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status": false,
			"msg":    *req.FilePath + " is a directory, not allow to delete directory",
		})
	}

	err = os.Remove(file)
	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status": false,
			"msg":    err.Error(),
		})
	}

	if AppCnf.Compress {
		// silently delete compressed file
		_ = os.Remove(file + ".fiber.gz")
	}

	if AppCnf.DeleteEmptyDir {
		dir := strings.Replace(file, "/"+f.Name(), "", 1)
		if dir != AppCnf.Path {
			empty, err := IsDirEmpty(dir)
			if err == nil && empty {
				_ = os.Remove(dir)
			}
		}
	}

	return c.JSON(fiber.Map{
		"status": true,
		"msg":    "file deleted",
	})
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

	if err = out.Validate(jwt.Expected{Issuer: AppCnf.ApiKey, Time: time.Now().UTC()}); err != nil {
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
