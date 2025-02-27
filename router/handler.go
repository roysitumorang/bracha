package router

import (
	"context"
	"encoding/gob"
	"errors"
	"net/url"
	"os"

	"github.com/roysitumorang/bracha/helper"
	serviceSadia "github.com/roysitumorang/bracha/services/sadia"
	"go.uber.org/zap"
)

type (
	Service struct {
		ServiceSadia *serviceSadia.ServiceSadia
	}
)

func MakeHandler(ctx context.Context) (*Service, error) {
	ctxt := "Router-MakeHandler"
	envSadiaBaseURL, ok := os.LookupEnv("SADIA_BASE_URL")
	if !ok || envSadiaBaseURL == "" {
		return nil, errors.New("env SADIA_BASE_URL is required")
	}
	sadiaURL, err := url.Parse(envSadiaBaseURL)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrParse")
		return nil, err
	}
	gob.Register(serviceSadia.User{})
	serviceSadia := serviceSadia.New(sadiaURL)
	return &Service{
		ServiceSadia: serviceSadia,
	}, nil
}
