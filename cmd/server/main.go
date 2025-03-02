package main

import (
	"fmt"
	"log/slog"
	"os"
	"runner/internal/app/handler"

	"github.com/gofiber/fiber/v2"
)

const PORT = 2111

func main() {
	err := os.MkdirAll("files", 0755)
	if err != nil {
		slog.Error("failed to create `./files/` directory", slog.Any("error", err))
		return
	}

	app := fiber.New()

	// NOTE: Sinve I don't give a shit about other content types here -
	// `application/json` will be set automatically, as I don't want to
	// set this header on every request
	app.Use(func(c *fiber.Ctx) error {
		c.Request().Header.Set("Content-Type", "application/json")
		return c.Next()
	})

	app.Get("/healthcheck", handler.Healthcheck)

	app.Post("/run", handler.RunSolution)
	app.Post("/test", handler.TestSolution)

	app.Listen(fmt.Sprintf(":%d", PORT))
}
