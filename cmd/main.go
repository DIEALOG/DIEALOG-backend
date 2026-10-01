package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/api/route"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/config"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/infra/database"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/infra/repository"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/usecase"
)

// @title          Fiber Backend API
// @version        1.0
// @description    Fiber Backend API
// @host			localhost:8080
// @BasePath		/api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token with the `Bearer ` prefix.
func main() {
	config.NewEnv()

	app := fiber.New(fiber.Config{
		AppName:      "Fiber Ent Clean Architecture",
		ServerHeader: "Fiber",
	})

	// Use global middlewares.
	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(etag.New())
	app.Use(favicon.New())
	app.Use(limiter.New(limiter.Config{
		Max: 100,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(&fiber.Map{
				"status":  "fail",
				"message": "You have requested too many in a single time-frame! Please wait another minute!",
			})
		},
	}))
	app.Use(logger.New())
	app.Use(requestid.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	api := app.Group("/api")

	app.Get("/swagger/*", swaggo.HandlerDefault)
	app.Get("/docs/*", swaggo.HandlerDefault)

	client, err := database.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if err = client.Schema.Create(context.Background()); err != nil {
		panic(err)
	}

	timeout := time.Duration(config.E.ContextTimeout) * time.Second

	// repository
	userRepository := repository.NewUserRepository(client)
	taskRepository := repository.NewTaskRepository(client)

	// usecase
	profileUsecase := usecase.NewProfileUseCase(userRepository, timeout)
	authUseCase := usecase.NewAuthUseCase(userRepository, timeout)
	taskUsecase := usecase.NewTaskUseCase(taskRepository, timeout)

	// router
	route.NewLoginRouter(api.Group("/login"), authUseCase)
	route.NewProfileRouter(api.Group("/profile"), profileUsecase)
	route.NewRefreshTokenRouter(api.Group("/refresh"), authUseCase)
	route.NewSignupRouter(api.Group("/signup"), authUseCase)
	route.NewTaskRouter(api.Group("/task"), taskUsecase)

	app.All("*", func(c fiber.Ctx) error {
		notFoundErr := fmt.Errorf(
			"route '%s' does not exist in this API",
			c.OriginalURL(),
		)

		return c.Status(http.StatusNotFound).JSON(&fiber.Map{
			"status": "error",
			"error":  notFoundErr.Error(),
		})
	})

	log.Fatal(app.Listen(config.E.ServerAddress))
}
