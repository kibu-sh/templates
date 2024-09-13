//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/foreman"
)

func InitServer() (*foreman.Manager, error) {
	wire.Build(wireSet)
	return nil, nil
}
