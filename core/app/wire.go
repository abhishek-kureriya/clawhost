//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/yourusername/clawhost/core/api"
	"github.com/yourusername/clawhost/core/logging"
)

func InitializeCoreServer(port string) (*api.CoreServer, error) {
	wire.Build(
		logging.NewLogger,
		api.NewRouter,
		api.NewCoreServer,
	)
	return nil, nil
}
