package internal

import (
	"os"
	"path"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func HandleDownloadFile(c *fiber.Ctx) error {
	token := c.Params("token")

	if len(token) == 0 {
		return sendResponse(c, fiber.StatusUnauthorized, false, "token require or invalid url")
	}

	out, err := verifyToken(token)
	if err != nil {
		return sendResponse(c, fiber.StatusUnauthorized, false, err.Error())
	}

	filePath := path.Join(AppCnf.Path, out.Subject)
	file, err := os.Lstat(filePath)
	if err != nil {
		ms := strings.SplitN(err.Error(), "/", -1)
		return sendResponse(c, fiber.StatusNotFound, false, ms[len(ms)-1])
	}

	c.Attachment(file.Name())
	return c.SendFile(filePath, AppCnf.Compress)
}

func HandleServeFile(c *fiber.Ctx) error {
	token := c.Params("token")

	if len(token) == 0 {
		return sendResponse(c, fiber.StatusUnauthorized, false, "token require or invalid url")
	}

	out, err := verifyToken(token)
	if err != nil {
		return sendResponse(c, fiber.StatusUnauthorized, false, err.Error())
	}

	c.Attachment(path.Base(out.Subject))
	c.Set("X-Accel-Redirect", path.Join(AppCnf.NginxFileServePath, out.Subject))
	return c.SendStatus(fiber.StatusOK)
}
