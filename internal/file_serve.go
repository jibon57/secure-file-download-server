package internal

import (
	"fmt"
	"os"
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

	file := fmt.Sprintf("%s/%s", AppCnf.Path, out.Subject)
	_, err = os.Lstat(file)

	if err != nil {
		ms := strings.SplitN(err.Error(), "/", -1)
		return sendResponse(c, fiber.StatusNotFound, false, ms[len(ms)-1])
	}

	c.Attachment(file)
	return c.SendFile(file, AppCnf.Compress)
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
	ms := strings.SplitN(out.Subject, "/", -1)

	c.Attachment(ms[len(ms)-1])
	c.Set("X-Accel-Redirect", fmt.Sprintf("%s%s", AppCnf.NginxFileServePath, out.Subject))
	return c.SendStatus(fiber.StatusOK)
}
