package internal

import (
	"os"
	"path"

	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofiber/fiber/v2"
)

func HandleDownloadFile(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*jwt.Claims)

	filePath := path.Join(AppCnf.Path, claims.Subject)
	file, err := os.Lstat(filePath)
	if err != nil {
		return sendResponse(c, fiber.StatusNotFound, false, "file not found")
	}

	c.Attachment(file.Name())
	return c.SendFile(filePath, AppCnf.Compress)
}

func HandleServeFile(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*jwt.Claims)

	c.Attachment(path.Base(claims.Subject))
	c.Set("X-Accel-Redirect", path.Join(AppCnf.NginxFileServePath, claims.Subject))
	return c.SendStatus(fiber.StatusOK)
}
