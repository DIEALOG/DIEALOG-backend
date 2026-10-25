package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/api/route"
	"github.com/DIEALOG/DIEALOG-backend/config"
	_ "github.com/DIEALOG/DIEALOG-backend/docs"
	"github.com/DIEALOG/DIEALOG-backend/infra/client/crypto"
	"github.com/DIEALOG/DIEALOG-backend/infra/database"
	"github.com/DIEALOG/DIEALOG-backend/infra/repository"
	"github.com/DIEALOG/DIEALOG-backend/internal/httpclient"
	"github.com/DIEALOG/DIEALOG-backend/usecase"
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
	"github.com/gofiber/fiber/v3/middleware/static"
)

// @title          DIEALOG Backend API
// @version        1.0
// @description    DIEALOG 디지털 유산 관리 서비스 백엔드 API
// @host			localhost:8080
// @BasePath		/api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token with the `Bearer ` prefix.
// @securityDefinitions.apikey ConfirmerAuth
// @in header
// @name X-Confirmer-Token
// @description Enter the confirmer token returned when an invitation is accepted.
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

	app.Get("/swagger", func(c fiber.Ctx) error {
		return c.Redirect().To("/swagger/index.html")
	})
	app.Get("/swagger/*", swaggo.HandlerDefault)
	app.Get("/docs", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs/index.html")
	})
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

	// client
	httpClient := httpclient.NewHTTPClient()
	cryptoClient := crypto.NewClient(httpClient, config.E.InternalSecretKey)

	// repository
	userRepository := repository.NewUserRepository(client)
	legacyRepository := repository.NewLegacyRepository(client)
	stakeholderRepository := repository.NewStakeholderRepository(client)
	stakeholderGroupRepository := repository.NewStakeholderGroupRepository(client)
	lastWordRepository := repository.NewLastWordRepository(client)
	userConfigRepository := repository.NewUserConfigRepository(client)
	transitionPolicyRepository := repository.NewUserTransitionPolicyRepository(client)
	transitionConfirmerRepository := repository.NewUserTransitionConfirmerRepository(client)
	aliveCheckRepository := repository.NewUserAliveCheckRepository(client)
	transitionCaseRepository := repository.NewUserTransitionCaseRepository(client)
	transitionConfirmationRepository := repository.NewUserTransitionConfirmationRepository(client)
	accountTaskRepository := repository.NewAccountHandlingTaskRepository(client)
	accountEventRepository := repository.NewAccountHandlingEventRepository(client)

	// usecase
	profileUseCase := usecase.NewProfileUseCase(userRepository, userConfigRepository, timeout)
	authUseCase := usecase.NewAuthUseCase(userRepository, userConfigRepository, timeout)
	credentialUseCase := usecase.NewCredentialUseCase(cryptoClient)
	legacyUseCase := usecase.NewLegacyUseCase(legacyRepository, credentialUseCase, timeout)
	stakeholderUseCase := usecase.NewStakeholderUseCase(stakeholderRepository, timeout)
	stakeholderGroupUseCase := usecase.NewStakeholderGroupUseCase(stakeholderGroupRepository, timeout)
	lastWordUseCase := usecase.NewLastWordUseCase(lastWordRepository, timeout)
	transitionPolicyUseCase := usecase.NewUserTransitionPolicyUseCase(transitionPolicyRepository, userConfigRepository, timeout)
	transitionConfirmerUseCase := usecase.NewUserTransitionConfirmerUseCase(transitionConfirmerRepository, stakeholderRepository, timeout)
	aliveCheckUseCase := usecase.NewUserAliveCheckUseCase(aliveCheckRepository, transitionPolicyUseCase, timeout)
	transitionCaseUseCase := usecase.NewUserTransitionCaseUseCase(transitionCaseRepository, transitionConfirmationRepository, transitionPolicyUseCase, timeout)
	transitionConfirmationUseCase := usecase.NewUserTransitionConfirmationUseCase(transitionConfirmationRepository, transitionCaseRepository, transitionCaseUseCase, timeout)
	accountHandlingUseCase := usecase.NewAccountHandlingUseCase(
		accountTaskRepository,
		accountEventRepository,
		legacyRepository,
		transitionCaseUseCase,
		timeout,
	)

	// router
	route.NewLoginRouter(api.Group("/login"), authUseCase)
	route.NewProfileRouter(api.Group("/profile"), profileUseCase)
	route.NewRefreshTokenRouter(api.Group("/refresh"), authUseCase)
	route.NewSignupRouter(api.Group("/signup"), authUseCase)
	route.NewLegacyRouter(api.Group("/legacy"), legacyUseCase, stakeholderUseCase, stakeholderGroupUseCase)
	route.NewLastWordRouter(api.Group("/last-word"), lastWordUseCase, stakeholderUseCase, stakeholderGroupUseCase)
	route.NewStakeholderRouter(api.Group("/stakeholder"), stakeholderUseCase, stakeholderGroupUseCase)
	route.NewTransitionRouter(
		api.Group("/transition"),
		transitionPolicyUseCase,
		transitionConfirmerUseCase,
		aliveCheckUseCase,
		transitionCaseUseCase,
		transitionConfirmationUseCase,
		accountHandlingUseCase,
	)

	// Static files must be registered before the catch-all 404 route.
	app.Get("/public/*", static.New("./public"))

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
