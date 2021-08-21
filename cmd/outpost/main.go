package main

import (
	"github.com/jointable/table/cmd/outpost/routes"
	"github.com/jointable/table/pkg/service"
)

func main() {
	s := service.New()

	if err := routes.Up(s); err != nil {
		panic(err)
	}

	panic(s.Run("localhost:9090"))
}
