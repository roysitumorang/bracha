package presenter

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/roysitumorang/bracha/helper"
	"github.com/roysitumorang/bracha/models"
	serviceSadia "github.com/roysitumorang/bracha/services/sadia"
	"go.uber.org/zap"
)

type (
	accountHTTPHandler struct {
		sessionStore *session.Store
		serviceSadia *serviceSadia.ServiceSadia
	}
)

func New(
	sessionStore *session.Store,
	serviceSadia *serviceSadia.ServiceSadia,
) *accountHTTPHandler {
	return &accountHTTPHandler{
		sessionStore: sessionStore,
		serviceSadia: serviceSadia,
	}
}

func (q *accountHTTPHandler) Mount(r fiber.Router) {
	r.Get("/logout", q.logout)
	r.Group("/login").
		Get("", q.login).
		Post("", q.doLogin)
	r.Group("/me").
		Get("/about", q.aboutCurrentUser)
}

func (q *accountHTTPHandler) logout(c *fiber.Ctx) error {
	ctxt := "AccountPresenter-logout"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if err = session.Destroy(); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrDestroy")
		return c.SendString(err.Error())
	}
	c.ClearCookie()
	return c.Redirect("/account/login")
}

func (q *accountHTTPHandler) login(c *fiber.Ctx) error {
	ctxt := "AccountPresenter-login"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if isAuthenticated, ok := session.Get(models.IsAuthenticated).(bool); ok && isAuthenticated {
		return c.Redirect("/account/me/about")
	}
	return c.Render("account/login", fiber.Map{
		"message": "",
		"login":   "",
	})
}

func (q *accountHTTPHandler) doLogin(c *fiber.Ctx) error {
	ctxt := "AccountPresenter-doLogin"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if isAuthenticated, ok := session.Get(models.IsAuthenticated).(bool); ok && isAuthenticated {
		return c.Redirect("/account/me/about")
	}
	response, err := q.serviceSadia.Login(ctx, c.FormValue("login"), c.FormValue("password"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrLogin")
		return c.Render("account/login", fiber.Map{
			"message": err.Error(),
			"login":   c.FormValue("login"),
		})
	}
	if response.StatusCode != fiber.StatusCreated {
		return c.Render("account/login", fiber.Map{
			"message": response.Message,
			"login":   c.FormValue("login"),
		})
	}
	session.Set(models.IsAuthenticated, true)
	session.Set(models.CurrentUser, response.Data.User)
	session.Set(models.CurrentJwt, response.Data.IDToken)
	session.SetExpiry(time.Until(response.Data.ExpiredAt))
	if err = session.Save(); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSave")
		return c.SendString(err.Error())
	}
	return c.Redirect("/account/me/about")
}

func (q *accountHTTPHandler) aboutCurrentUser(c *fiber.Ctx) error {
	ctxt := "AccountPresenter-aboutCurrentUser"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if isAuthenticated, ok := session.Get(models.IsAuthenticated).(bool); !ok || !isAuthenticated {
		return c.Redirect("/account/login")
	}
	currentUser, ok := session.Get(models.CurrentUser).(serviceSadia.User)
	if !ok {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Render("account/me/about", fiber.Map{
		"currentUser": currentUser,
	})
}
