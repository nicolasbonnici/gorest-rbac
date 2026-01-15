package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest-rbac"
)

type Article struct {
	ID      int    `json:"id" rbac:"read:*;write:none"`
	Title   string `json:"title" rbac:"read:*;write:editor,admin"`
	Content string `json:"content" rbac:"read:*;write:editor,admin"`
	Status  string `json:"status" rbac:"read:editor,admin;write:admin"`
	Author  string `json:"author" rbac:"read:*;write:none"`
}

func main() {
	app := fiber.New()

	cfg := rbac.Config{
		DefaultPolicy: rbac.DenyAll,
		SuperuserRole: "admin",
		RoleHierarchy: map[string][]string{
			"admin":  {"editor"},
			"editor": {"user"},
		},
		CacheEnabled:       true,
		CacheTTL:           300,
		StrictMode:         true,
		DefaultFieldPolicy: "deny",
	}

	voter, err := rbac.NewVoter(cfg)
	if err != nil {
		log.Fatalf("Failed to create voter: %v", err)
	}

	roleProvider := rbac.NewFiberRoleProvider("user_roles", "user_id")

	app.Use(rbac.Middleware(voter, roleProvider))

	app.Get("/articles/:id", func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		article := Article{
			ID:      1,
			Title:   "Introduction to RBAC",
			Content: "Role-Based Access Control is a method of restricting system access...",
			Status:  "published",
			Author:  "John Doe",
		}

		filtered, err := voter.FilterRead(ctx, &article)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(filtered)
	})

	app.Post("/articles", func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		var article Article
		if err := c.BodyParser(&article); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if err := voter.ValidateWrite(ctx, &article); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Article created successfully",
			"article": article,
		})
	})

	log.Fatal(app.Listen(":3000"))
}
