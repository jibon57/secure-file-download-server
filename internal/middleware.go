package internal

import (
	"github.com/gofiber/fiber/v2"
)

func VerifyTokenMiddleware(c *fiber.Ctx) error {
	token := c.Params("token")

	if len(token) == 0 {
		return sendResponse(c, fiber.StatusUnauthorized, false, "token require or invalid url")
	}

	out, err := verifyToken(token)
	if err != nil {
		return sendResponse(c, fiber.StatusUnauthorized, false, err.Error())
	}

	c.Locals("claims", out)
	return c.Next()
}
