package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
	})
}

func InternalServerError(c *fiber.Ctx) error {
	return Error(c, http.StatusInternalServerError, "internal server error")
}
