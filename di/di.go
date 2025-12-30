package di

import (
	_ "github.com/lib/pq" // <------------ here
	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/ome/controller"
	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/ome/service"
	"github.com/toufiq-austcse/go-api-boilerplate/pkg/db/providers/mongodb"
	"github.com/toufiq-austcse/go-api-boilerplate/pkg/http_clients"
	"go.uber.org/dig"
)

func NewDiContainer() (*dig.Container, error) {
	c := dig.New()
	providers := []interface {
	}{
		controller.NewOmeController,
		http_clients.NewOmeHTTPClient,
		mongodb.New,
		service.NewOmeService,
	}
	for _, provider := range providers {
		if err := c.Provide(provider); err != nil {
			return nil, err
		}
	}
	return c, nil
}
