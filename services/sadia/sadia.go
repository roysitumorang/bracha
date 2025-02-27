package sadia

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/roysitumorang/bracha/helper"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	ServiceSadia struct {
		baseURL *url.URL
	}

	LoginRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	UserLoginResponse struct {
		IDToken   string    `json:"id_token"`
		ExpiredAt time.Time `json:"expired_at"`
		User      User      `json:"user"`
	}

	User struct {
		ID                      string     `json:"id"`
		AccountType             uint8      `json:"account_type"`
		Status                  int8       `json:"status"`
		Name                    string     `json:"name"`
		Username                string     `json:"username"`
		ConfirmedAt             *time.Time `json:"confirmed_at"`
		Email                   *string    `json:"email"`
		UnconfirmedEmail        *string    `json:"unconfirmed_email"`
		EmailConfirmationSentAt *time.Time `json:"email_confirmation_sent_at"`
		EmailConfirmedAt        *time.Time `json:"email_confirmed_at"`
		Phone                   *string    `json:"phone"`
		UnconfirmedPhone        *string    `json:"unconfirmed_phone"`
		PhoneConfirmationSentAt *time.Time `json:"phone_confirmation_sent_at"`
		PhoneConfirmedAt        *time.Time `json:"phone_confirmed_at"`
		LastPasswordChange      *time.Time `json:"last_password_change"`
		ResetPasswordSentAt     *time.Time `json:"reset_password_sent_at"`
		LoginCount              uint       `json:"login_count"`
		CurrentLoginAt          *time.Time `json:"current_login_at"`
		CurrentLoginIP          *string    `json:"current_login_ip"`
		LastLoginAt             *time.Time `json:"last_login_at"`
		LastLoginIP             *string    `json:"last_login_ip"`
		LoginFailedAttempts     int        `json:"login_failed_attempts"`
		LoginLockedAt           *time.Time `json:"login_locked_at"`
		CreatedAt               time.Time  `json:"created_at"`
		UpdatedAt               time.Time  `json:"updated_at"`
		DeactivatedAt           *time.Time `json:"deactivated_at"`
		CompanyID               string     `json:"company_id"`
		UserLevel               uint8      `json:"user_level"`
		CurrentSessionID        *string    `json:"current_session_id"`
	}

	ResponseUserLogin struct {
		RequestID  string            `json:"request_id"`
		RequestURL string            `json:"request_url"`
		StatusCode int               `json:"status_code"`
		Status     string            `json:"status"`
		Message    string            `json:"message"`
		Timestamp  time.Time         `json:"timestamp"`
		Latency    string            `json:"latency"`
		App        string            `json:"app"`
		Data       UserLoginResponse `json:"data"`
	}

	Pagination struct {
		Links struct {
			First    string `json:"first"`
			Previous string `json:"previous"`
			Current  string `json:"current"`
			Next     string `json:"next"`
		} `json:"links"`
		Info struct {
			Limit int64 `json:"limit"`
			Pages int64 `json:"pages"`
			Total int64 `json:"total"`
		} `json:"info"`
	}

	ProductCategory struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	ProductCategoriesPagination struct {
		Pagination Pagination        `json:"pagination"`
		Rows       []ProductCategory `json:"rows"`
	}

	ResponseProductCategoriesPagination struct {
		RequestID  string                      `json:"request_id"`
		RequestURL string                      `json:"request_url"`
		StatusCode int                         `json:"status_code"`
		Status     string                      `json:"status"`
		Message    string                      `json:"message"`
		Timestamp  time.Time                   `json:"timestamp"`
		Latency    string                      `json:"latency"`
		App        string                      `json:"app"`
		Data       ProductCategoriesPagination `json:"data"`
	}

	ResponseProductCategory struct {
		RequestID  string          `json:"request_id"`
		RequestURL string          `json:"request_url"`
		StatusCode int             `json:"status_code"`
		Status     string          `json:"status"`
		Message    string          `json:"message"`
		Timestamp  time.Time       `json:"timestamp"`
		Latency    string          `json:"latency"`
		App        string          `json:"app"`
		Data       ProductCategory `json:"data"`
	}
)

var (
	errInvalidJSON = errors.New("invalid JSON payload received")
)

func New(baseURL *url.URL) *ServiceSadia {
	return &ServiceSadia{
		baseURL: baseURL,
	}
}

func (q *ServiceSadia) Login(ctx context.Context, login, password string) (*ResponseUserLogin, error) {
	ctxt := "ServiceSadia-Login"
	request := LoginRequest{
		Login:    login,
		Password: helper.Base64Encode(password),
	}
	_, statusCode, respBody, err := q.hitEndpoint(ctx, "/account/login", fiber.MethodPost, nil, "", request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrHitEndpoint")
		return nil, err
	}
	if !json.Valid(respBody) {
		return nil, errInvalidJSON
	}
	var response ResponseUserLogin
	if err = json.Unmarshal(respBody, &response); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	if statusCode != fiber.StatusCreated {
		return nil, errors.New(response.Message)
	}
	return &response, nil
}

func (q *ServiceSadia) FindProductCategories(ctx context.Context, jwt string, urlValues url.Values) (*ResponseProductCategoriesPagination, error) {
	ctxt := "ServiceSadia-FindProductCategories"
	_, statusCode, respBody, err := q.hitEndpoint(ctx, "/product_category", fiber.MethodGet, urlValues, jwt)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrHitEndpoint")
		return nil, err
	}
	if !json.Valid(respBody) {
		return nil, errInvalidJSON
	}
	var response ResponseProductCategoriesPagination
	if err = json.Unmarshal(respBody, &response); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	if statusCode != fiber.StatusOK {
		return nil, errors.New(response.Message)
	}
	return &response, nil
}

func (q *ServiceSadia) CreateProductCategory(ctx context.Context, jwt, name, slug string) (*ResponseProductCategory, error) {
	ctxt := "ServiceSadia-CreateProductCategory"
	request := ProductCategory{
		Name: name,
		Slug: slug,
	}
	_, statusCode, respBody, err := q.hitEndpoint(ctx, "/product_category", fiber.MethodPost, nil, jwt, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrHitEndpoint")
		return nil, err
	}
	if !json.Valid(respBody) {
		return nil, errInvalidJSON
	}
	var response ResponseProductCategory
	if err = json.Unmarshal(respBody, &response); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	if statusCode != fiber.StatusCreated {
		return nil, errors.New(response.Message)
	}
	return &response, nil
}

func (q *ServiceSadia) FindProductCategory(ctx context.Context, jwt, productCategoryID string) (*ResponseProductCategory, error) {
	ctxt := "ServiceSadia-FindProductCategory"
	var builder strings.Builder
	_, _ = builder.WriteString("/product_category/")
	_, _ = builder.WriteString(productCategoryID)
	_, statusCode, respBody, err := q.hitEndpoint(ctx, builder.String(), fiber.MethodGet, nil, jwt)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrHitEndpoint")
		return nil, err
	}
	if !json.Valid(respBody) {
		return nil, errInvalidJSON
	}
	var response ResponseProductCategory
	if err = json.Unmarshal(respBody, &response); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	if statusCode != fiber.StatusOK {
		return nil, errors.New(response.Message)
	}
	return &response, nil
}

func (q *ServiceSadia) UpdateProductCategory(ctx context.Context, jwt, productCategoryID, name, slug string) (*ResponseProductCategory, error) {
	ctxt := "ServiceSadia-UpdateProductCategory"
	var builder strings.Builder
	_, _ = builder.WriteString("/product_category/")
	_, _ = builder.WriteString(productCategoryID)
	request := ProductCategory{
		Name: name,
		Slug: slug,
	}
	_, statusCode, respBody, err := q.hitEndpoint(ctx, builder.String(), fiber.MethodPut, nil, jwt, request)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrHitEndpoint")
		return nil, err
	}
	if !json.Valid(respBody) {
		return nil, errInvalidJSON
	}
	var response ResponseProductCategory
	if err = json.Unmarshal(respBody, &response); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUnmarshal")
		return nil, err
	}
	if statusCode != fiber.StatusOK {
		return nil, errors.New(response.Message)
	}
	return &response, nil
}

func (q *ServiceSadia) hitEndpoint(ctx context.Context, endpoint, requestMethod string, urlValues url.Values, jwt string, payload ...any) (requestURL string, statusCode int, responseBody []byte, err error) {
	ctxt := "ServiceSadia-hitEndpoint"
	var builder strings.Builder
	_, _ = builder.WriteString(q.baseURL.String())
	_, _ = builder.WriteString(endpoint)
	request := fasthttp.AcquireRequest()
	request.SetRequestURI(builder.String())
	request.Header.SetMethod(requestMethod)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set(fiber.HeaderXRequestID, uuid.New().String())
	if queryString := urlValues.Encode(); queryString != "" {
		request.URI().SetQueryString(queryString)
		_, _ = builder.WriteString("?")
		_, _ = builder.WriteString(queryString)
	}
	requestURL = builder.String()
	if jwt != "" {
		builder.Reset()
		_, _ = builder.WriteString("Bearer ")
		_, _ = builder.WriteString(jwt)
		request.Header.Set(fiber.HeaderAuthorization, builder.String())
	}
	if requestMethod != fiber.MethodGet &&
		len(payload) > 0 && payload[0] != nil {
		requestBody, err := json.Marshal(payload[0])
		if err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrMarshal")
			return requestURL, 0, nil, err
		}
		request.SetBody(requestBody)
	}
	response := fasthttp.AcquireResponse()
	if err = fasthttp.Do(request, response); err != nil {
		for errors.Unwrap(err) != nil {
			err = errors.Unwrap(err)
		}
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrDo")
		return
	}
	fasthttp.ReleaseRequest(request)
	statusCode = response.StatusCode()
	responseBody = response.Body()
	builder.Reset()
	_, _ = builder.WriteString(requestMethod)
	_, _ = builder.WriteString(" ")
	_, _ = builder.WriteString(requestURL)
	_, _ = builder.WriteString(" ")
	_, _ = builder.WriteString(strconv.Itoa(statusCode))
	helper.Log(ctx, zap.InfoLevel, builder.String(), ctxt, "")
	return
}
