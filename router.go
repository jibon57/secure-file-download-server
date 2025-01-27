package main

import (
	"errors"
	"fmt"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"io/fs"
	"log"
	"os"
	"path"
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
		return sendResponse(c, fiber.StatusUnauthorized, false, "token require or invalid url")
	}

	file, status, err := verifyToken(token)
	if err != nil {
		return sendResponse(c, status, false, err.Error())
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
		return sendResponse(c, fiber.StatusUnauthorized, false, "Auth header information are missing")
	}

	if apiKey != AppCnf.ApiKey || secret != AppCnf.ApiSecret {
		return sendResponse(c, fiber.StatusUnauthorized, false, "Auth header information didn't match")
	}

	req := new(DeleteFileReq)
	err := c.BodyParser(req)
	if err != nil {
		return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
	}

	if req.FilePath == nil {
		return sendResponse(c, fiber.StatusNotAcceptable, false, "file_path value require")
	}

	file := fmt.Sprintf("%s/%s", AppCnf.Path, *req.FilePath)
	f, err := os.Stat(file)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			return sendResponse(c, fiber.StatusNotFound, false, *req.FilePath+" does not exist")
		} else {
			return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
		}
	}

	if f.IsDir() {
		return sendResponse(c, fiber.StatusNotAcceptable, false, *req.FilePath+" is a directory, not allow to delete directory")
	}

	if AppCnf.EnableDelFileBackup {
		// first with the video file
		toFile := path.Join(AppCnf.DelFileBackupPath, f.Name())
		err := os.Rename(file, toFile)
		if err != nil {
			log.Println(err)
			return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
		}

		// otherwise during cleanup will be hard to detect
		newTime := time.Now()
		if err := os.Chtimes(toFile, newTime, newTime); err != nil {
			log.Println("Failed to update file modification time:", err)
		}
	} else {
		err = os.Remove(file)
		if err != nil {
			return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
		}
	}

	if AppCnf.Compress {
		// silently delete compressed file
		_ = os.Remove(file + ".fiber.gz")
	}

	if AppCnf.DeleteEmptyDir {
		dir := strings.Replace(file, "/"+f.Name(), "", 1)
		if dir != AppCnf.Path {
			if empty, err := IsDirEmpty(dir); err == nil && empty {
				_ = os.Remove(dir)
			}
		}
	}

	return sendResponse(c, fiber.StatusOK, true, "file deleted")
}

func verifyToken(token string) (string, int, error) {
	tok, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.HS256})
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

func sendResponse(c *fiber.Ctx, statusCode int, status bool, msg string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status": status,
		"msg":    msg,
	})
}
