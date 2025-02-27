package router

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/rewrite"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/valkey"
	"github.com/gofiber/template/jet/v2"
	"github.com/joho/godotenv"
	"github.com/roysitumorang/bracha/config"
	_ "github.com/roysitumorang/bracha/docs"
	"github.com/roysitumorang/bracha/helper"
	"github.com/roysitumorang/bracha/middleware"
	accountPresenter "github.com/roysitumorang/bracha/modules/account/presenter"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"go.uber.org/zap"
)

const (
	DefaultPort uint16 = 8080
)

func (q *Service) HTTPServerMain(ctx context.Context) error {
	ctxt := "Router-HTTPServerMain"
	// Create a new engine
	engine := jet.New("./views", ".jet")
	storage := valkey.New(valkey.Config{
		URL: os.Getenv("REDIS_URL"),
	})
	sessionStore := session.New(session.Config{
		Storage: storage,
	})
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		Views:       engine,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			statusCode := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				statusCode = e.Code
			}
			return helper.NewResponse(statusCode).SetMessage(err.Error()).WriteResponse(ctx)
		},
	})
	app.Use(
		recover.New(recover.Config{
			EnableStackTrace: true,
		}),
		fiberzap.New(fiberzap.Config{
			Logger: helper.GetLogger(),
		}),
		requestid.New(),
		compress.New(),
		rewrite.New(rewrite.Config{
			Rules: map[string]string{},
		}),
		cors.New(),
	)
	basicAuth := middleware.BasicAuth()
	if helper.GetEnv() == "development" {
		app.Get("/swagger/*", fiberSwagger.WrapHandler)
	}
	app.Get("/ping", func(c *fiber.Ctx) error {
		return helper.NewResponse(fiber.StatusOK).
			SetData(map[string]any{
				"version": config.Version,
				"commit":  config.Commit,
				"build":   config.Build,
				"upsince": config.Now.Format(time.RFC3339),
				"uptime":  time.Since(config.Now).String(),
			}).WriteResponse(c)
	}).
		Get("/metrics", basicAuth, monitor.New(monitor.Config{
			APIOnly: true,
		})).
		Get("/env", basicAuth, func(c *fiber.Ctx) error {
			envMap, err := godotenv.Read(".env")
			if err != nil {
				helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRead")
				return helper.NewResponse(fiber.StatusBadRequest).SetMessage(err.Error()).WriteResponse(c)
			}
			envMap["GO_VERSION"] = runtime.Version()
			return helper.NewResponse(fiber.StatusOK).SetData(envMap).WriteResponse(c)
		})
	accountPresenter.New(sessionStore, q.ServiceSadia).Mount(app.Group("/account"))
	app.Use(func(c *fiber.Ctx) error {
		return helper.NewResponse(fiber.StatusNotFound).WriteResponse(c)
	})
	port := DefaultPort
	if envPort, ok := os.LookupEnv("PORT"); ok && envPort != "" {
		if portInt, _ := strconv.Atoi(envPort); portInt >= 0 && portInt <= math.MaxUint16 {
			port = uint16(portInt)
		}
	}
	listenerPort := fmt.Sprintf(":%d", port)
	err := app.Listen(listenerPort)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrListen")
	}
	return err
}
