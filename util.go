package main

import "github.com/gofiber/fiber/v2"

func jsonMessage(c *fiber.Ctx, msg string) error {
	return c.JSON(fiber.Map{"message": msg})
}

func jsonError(c *fiber.Ctx, msg string, status int) error {
	return c.Status(status).JSON(fiber.Map{"message": msg})
}

func isLoggedIn2(c *fiber.Ctx) bool {
	return db.SessionCheck(c)
}

func isLoggedOut2(c *fiber.Ctx) bool {
	return !isLoggedIn2(c)
}

func checkPasswordTry2(c *fiber.Ctx) error {
	if passwordTry >= passwordMaxTry {
		_ = db.Close()
		msg := "No more try. Input wrong password too many times."
		return jsonError(c, msg, 403)
	}
	return nil
}
