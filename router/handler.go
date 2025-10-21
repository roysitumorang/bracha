package router

import (
	"context"
	"encoding/gob"

	"github.com/roysitumorang/bracha/helper"
	serviceSadia "github.com/roysitumorang/bracha/services/sadia"
)

type (
	Service struct {
		ServiceSadia *serviceSadia.ServiceSadia
	}
)

func MakeHandler(ctx context.Context) (*Service, error) {
	gob.Register(serviceSadia.User{})
	serviceSadia := serviceSadia.New(helper.GetSadiaBaseURL())
	return &Service{
		ServiceSadia: serviceSadia,
	}, nil
}
