package presenter

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/roysitumorang/bracha/helper"
	"github.com/roysitumorang/bracha/models"
	serviceSadia "github.com/roysitumorang/bracha/services/sadia"
	"go.uber.org/zap"
)

type (
	productCategoryHTTPHandler struct {
		sessionStore *session.Store
		serviceSadia *serviceSadia.ServiceSadia
	}
)

func New(
	sessionStore *session.Store,
	serviceSadia *serviceSadia.ServiceSadia,
) *productCategoryHTTPHandler {
	return &productCategoryHTTPHandler{
		sessionStore: sessionStore,
		serviceSadia: serviceSadia,
	}
}

func (q *productCategoryHTTPHandler) Mount(r fiber.Router) {
	r.Get("", q.index).
		Get("/new", q.new).
		Post("", q.create).
		Get("/:id/edit", q.edit).
		Post("/:id", q.update)
}

func (q *productCategoryHTTPHandler) index(c *fiber.Ctx) error {
	ctxt := "ProductCategoryPresenter-index"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if isAuthenticated, ok := session.Get(models.IsAuthenticated).(bool); !ok || !isAuthenticated {
		return c.Redirect("/account/login")
	}
	currentUser, currentUserOk := session.Get(models.CurrentUser).(serviceSadia.User)
	jwt, jwtOk := session.Get(models.CurrentJwt).(string)
	if !currentUserOk || !jwtOk {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	originalURL, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParseRequestURI")
		return c.SendString(err.Error())
	}
	response, err := q.serviceSadia.FindProductCategories(ctx, jwt, originalURL.Query())
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	if response == nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Render("product_category/index", fiber.Map{
		"is_authenticated": true,
		"currentUser":      currentUser,
		"message":          "",
		"response":         response,
	})
}

func (q *productCategoryHTTPHandler) new(c *fiber.Ctx) error {
	ctxt := "ProductCategoryPresenter-new"
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
	return c.Render("product_category/new", fiber.Map{
		"is_authenticated": true,
		"currentUser":      currentUser,
		"message":          "",
		"name":             "",
		"slug":             "",
	})
}

func (q *productCategoryHTTPHandler) create(c *fiber.Ctx) error {
	ctxt := "ProductCategoryPresenter-create"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if isAuthenticated, ok := session.Get(models.IsAuthenticated).(bool); !ok || !isAuthenticated {
		return c.Redirect("/account/login")
	}
	currentUser, currentUserOk := session.Get(models.CurrentUser).(serviceSadia.User)
	jwt, jwtOk := session.Get(models.CurrentJwt).(string)
	if !currentUserOk || !jwtOk {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if _, err = q.serviceSadia.CreateProductCategory(ctx, jwt, c.FormValue("name"), c.FormValue("slug")); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateProductCategory")
		return c.Render("product_category/new", fiber.Map{
			"is_authenticated": true,
			"currentUser":      currentUser,
			"message":          err.Error(),
			"name":             c.FormValue("name"),
			"slug":             c.FormValue("slug"),
		})
	}
	return c.Redirect("/product_category")
}

func (q *productCategoryHTTPHandler) edit(c *fiber.Ctx) error {
	ctxt := "ProductCategoryPresenter-edit"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if isAuthenticated, ok := session.Get(models.IsAuthenticated).(bool); !ok || !isAuthenticated {
		return c.Redirect("/account/login")
	}
	currentUser, currentUserOk := session.Get(models.CurrentUser).(serviceSadia.User)
	jwt, jwtOk := session.Get(models.CurrentJwt).(string)
	if !currentUserOk || !jwtOk {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	response, err := q.serviceSadia.FindProductCategory(ctx, jwt, c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategory")
		return c.SendString(err.Error())
	}
	if response == nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Render("product_category/edit", fiber.Map{
		"is_authenticated": true,
		"currentUser":      currentUser,
		"message":          "",
		"id":               response.Data.ID,
		"name":             response.Data.Name,
		"slug":             response.Data.Slug,
	})
}

func (q *productCategoryHTTPHandler) update(c *fiber.Ctx) error {
	ctxt := "ProductCategoryPresenter-update"
	ctx := c.UserContext()
	session, err := q.sessionStore.Get(c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrGet")
		return c.SendString(err.Error())
	}
	if isAuthenticated, ok := session.Get(models.IsAuthenticated).(bool); !ok || !isAuthenticated {
		return c.Redirect("/account/login")
	}
	currentUser, currentUserOk := session.Get(models.CurrentUser).(serviceSadia.User)
	jwt, jwtOk := session.Get(models.CurrentJwt).(string)
	if !currentUserOk || !jwtOk {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if _, err := q.serviceSadia.UpdateProductCategory(ctx, jwt, c.Params("id"), c.FormValue("name"), c.FormValue("slug")); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateProductCategory")
		return c.Render("product_category/edit", fiber.Map{
			"is_authenticated": true,
			"currentUser":      currentUser,
			"message":          err.Error(),
			"id":               c.Params("id"),
			"name":             c.FormValue("name"),
			"slug":             c.FormValue("slug"),
		})
	}
	return c.Redirect("/product_category")
}
