package main

import (
	"github.com/kibu-sh/kibu/pkg/dotenv"
)

func init() {
	dotenv.AutoLoadDotEnv(nil)
}

func main() {
	server, err := InitServer()
	if err != nil {
		panic(err)
	}

	if err = server.Wait(); err != nil {
		panic(err)
	}
}
