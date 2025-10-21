package presenter

import (
	"errors"
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
	originalURL, err := url.ParseRequestURI(helper.ByteSlice2String(c.Context().URI().FullURI()))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParseRequestURI")
		return c.SendString(err.Error())
	}
	response, err := q.serviceSadia.FindProducts(ctx, jwt, originalURL)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProducts")
		return c.SendString(err.Error())
	}
	if response == nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	return c.Render("product/index", fiber.Map{
		"is_authenticated": true,
		"currentUser":      currentUser,
		"message":          "",
		"q":                c.Query("q"),
		"limit":            limit,
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
	var product serviceSadia.Product
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
		"product":           product,
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
	var product serviceSadia.Product
	if err := c.BodyParser(&product); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBodyParser")
		return c.SendString(err.Error())
	}
	var categoryID string
	if product.CategoryID != nil {
		categoryID = *product.CategoryID
		if *product.CategoryID == "" {
			product.CategoryID = nil
		}
	}
	productCategories, err := q.serviceSadia.FindProductCategories(ctx, jwt, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	if product.Stock < 0 {
		err = errors.New("stock: requires a positive integer")
	} else if product.PurchasePrice < 0 {
		err = errors.New("purchase_price: requires a positive integer")
	} else if product.SellingPrice < 0 {
		err = errors.New("selling_price: requires a positive integer")
	}
	if err == nil {
		if _, err = q.serviceSadia.CreateProduct(ctx, jwt, product); err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateProduct")
		}
	}
	if err != nil {
		return c.Render("product/new", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           err.Error(),
			"productCategories": productCategories,
			"categoryID":        categoryID,
			"product":           product,
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
	product := response.Data
	var categoryID string
	if product.CategoryID != nil {
		categoryID = *product.CategoryID
	}
	productCategories, err := q.serviceSadia.FindProductCategories(ctx, jwt, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	return c.Render("product/edit", fiber.Map{
		"is_authenticated":  true,
		"currentUser":       currentUser,
		"message":           "",
		"productCategories": productCategories,
		"categoryID":        categoryID,
		"product":           product,
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
	response, err := q.serviceSadia.FindProduct(ctx, jwt, c.Params("id"))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProduct")
		return c.SendString(err.Error())
	}
	if response == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	var product serviceSadia.Product
	if err := c.BodyParser(&product); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBodyParser")
		return c.SendString(err.Error())
	}
	product.ID = response.Data.ID
	var categoryID string
	if product.CategoryID != nil {
		categoryID = *product.CategoryID
		if *product.CategoryID == "" {
			product.CategoryID = nil
		}
	}
	productCategories, err := q.serviceSadia.FindProductCategories(ctx, jwt, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindProductCategories")
		return c.SendString(err.Error())
	}
	if product.Stock < 0 {
		err = errors.New("stock: requires a positive integer")
	} else if product.PurchasePrice < 0 {
		err = errors.New("purchase_price: requires a positive integer")
	} else if product.SellingPrice < 0 {
		err = errors.New("selling_price: requires a positive integer")
	}
	if err == nil {
		if _, err = q.serviceSadia.UpdateProduct(ctx, jwt, product.ID, product); err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUpdateProduct")
		}
	}
	if err != nil {
		return c.Render("product/edit", fiber.Map{
			"is_authenticated":  true,
			"currentUser":       currentUser,
			"message":           err.Error(),
			"productCategories": productCategories,
			"categoryID":        categoryID,
			"product":           product,
		})
	}
	return c.Redirect("/product")
}
