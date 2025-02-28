package presenter

import (
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/roysitumorang/bracha/helper"
	"github.com/roysitumorang/bracha/models"
	serviceSadia "github.com/roysitumorang/bracha/services/sadia"
	"go.uber.org/zap"
)

type (
	productHTTPHandler struct {
		sessionStore *session.Store
		serviceSadia *serviceSadia.ServiceSadia
	}
)

func New(
	sessionStore *session.Store,
	serviceSadia *serviceSadia.ServiceSadia,
) *productHTTPHandler {
	return &productHTTPHandler{
		sessionStore: sessionStore,
		serviceSadia: serviceSadia,
	}
}

func (q *productHTTPHandler) Mount(r fiber.Router) {
	r.Get("", q.index).
		Get("/new", q.new).
		Post("", q.create).
		Get("/:id/edit", q.edit).
		Post("/:id", q.update)
}

func (q *productHTTPHandler) index(c *fiber.Ctx) error {
	ctxt := "ProductPresenter-index"
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
	response, err := q.serviceSadia.FindProducts(ctx, jwt, originalURL.Query())
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProducts")
		return c.SendString(err.Error())
	}
	if response == nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Render("product/index", fiber.Map{
		"is_authenticated": true,
		"currentUser":      currentUser,
		"message":          "",
		"response":         response,
	})
}

func (q *productHTTPHandler) new(c *fiber.Ctx) error {
	ctxt := "ProductPresenter-new"
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
	productCategories, err := q.serviceSadia.FindProductCategories(ctx, jwt, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	return c.Render("product/new", fiber.Map{
		"is_authenticated":  true,
		"currentUser":       currentUser,
		"message":           "",
		"productCategories": productCategories,
		"categoryID":        "",
		"name":              "",
		"slug":              "",
		"uom":               "",
		"stock":             0,
		"price":             0,
	})
}

func (q *productHTTPHandler) create(c *fiber.Ctx) error {
	ctxt := "ProductPresenter-create"
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
	productCategories, err := q.serviceSadia.FindProductCategories(ctx, jwt, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	stock, err := strconv.ParseInt(c.FormValue("stock"), 10, 64)
	if err != nil || stock < 0 {
		return c.Render("product/new", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           "stock: requires a positive integer",
			"productCategories": productCategories,
			"categoryID":        c.FormValue("category_id"),
			"name":              c.FormValue("name"),
			"slug":              c.FormValue("slug"),
			"uom":               c.FormValue("uom"),
			"stock":             c.FormValue("stock"),
			"price":             c.FormValue("price"),
		})
	}
	price, err := strconv.ParseInt(c.FormValue("price"), 10, 64)
	if err != nil || stock < 0 {
		return c.Render("product/new", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           "price: requires a positive integer",
			"productCategories": productCategories,
			"categoryID":        c.FormValue("category_id"),
			"name":              c.FormValue("name"),
			"slug":              c.FormValue("slug"),
			"uom":               c.FormValue("uom"),
			"stock":             c.FormValue("stock"),
			"price":             c.FormValue("price"),
		})
	}
	if _, err = q.serviceSadia.CreateProduct(ctx, jwt, c.FormValue("category_id"), c.FormValue("name"), c.FormValue("slug"), c.FormValue("uom"), stock, price); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateProduct")
		return c.Render("product/new", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           err.Error(),
			"productCategories": productCategories,
			"categoryID":        c.FormValue("category_id"),
			"name":              c.FormValue("name"),
			"slug":              c.FormValue("slug"),
			"uom":               c.FormValue("uom"),
			"stock":             c.FormValue("stock"),
			"price":             c.FormValue("price"),
		})
	}
	return c.Redirect("/product")
}

func (q *productHTTPHandler) edit(c *fiber.Ctx) error {
	ctxt := "ProductPresenter-edit"
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
	response, err := q.serviceSadia.FindProduct(ctx, jwt, c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProduct")
		return c.SendString(err.Error())
	}
	if response == nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	productCategories, err := q.serviceSadia.FindProductCategories(ctx, jwt, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	var categoryID string
	if response.Data.CategoryID != nil {
		categoryID = *response.Data.CategoryID
	}
	return c.Render("product/edit", fiber.Map{
		"is_authenticated":  true,
		"currentUser":       currentUser,
		"message":           "",
		"productCategories": productCategories,
		"id":                response.Data.ID,
		"categoryID":        categoryID,
		"name":              response.Data.Name,
		"slug":              response.Data.Slug,
		"uom":               response.Data.UOM,
		"stock":             response.Data.Stock,
		"price":             response.Data.Price,
	})
}

func (q *productHTTPHandler) update(c *fiber.Ctx) error {
	ctxt := "ProductPresenter-update"
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
	productCategories, err := q.serviceSadia.FindProductCategories(ctx, jwt, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	stock, err := strconv.ParseInt(c.FormValue("stock"), 10, 64)
	if err != nil || stock < 0 {
		return c.Render("product/edit", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           "stock: requires a positive integer",
			"productCategories": productCategories,
			"id":                c.Params("id"),
			"categoryID":        c.FormValue("category_id"),
			"name":              c.FormValue("name"),
			"slug":              c.FormValue("slug"),
			"uom":               c.FormValue("uom"),
			"stock":             c.FormValue("stock"),
			"price":             c.FormValue("price"),
		})
	}
	price, err := strconv.ParseInt(c.FormValue("price"), 10, 64)
	if err != nil || stock < 0 {
		return c.Render("product/edit", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           "price: requires a positive integer",
			"productCategories": productCategories,
			"id":                c.Params("id"),
			"categoryID":        c.FormValue("category_id"),
			"name":              c.FormValue("name"),
			"slug":              c.FormValue("slug"),
			"uom":               c.FormValue("uom"),
			"stock":             c.FormValue("stock"),
			"price":             c.FormValue("price"),
		})
	}
	if _, err := q.serviceSadia.UpdateProduct(ctx, jwt, c.Params("id"), c.FormValue("category_id"), c.FormValue("name"), c.FormValue("slug"), c.FormValue("uom"), stock, price); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateProduct")
		return c.Render("product/edit", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           err.Error(),
			"productCategories": productCategories,
			"id":                c.Params("id"),
			"categoryID":        c.FormValue("category_id"),
			"name":              c.FormValue("name"),
			"slug":              c.FormValue("slug"),
			"uom":               c.FormValue("uom"),
			"stock":             c.FormValue("stock"),
			"price":             c.FormValue("price"),
		})
	}
	return c.Redirect("/product")
}
