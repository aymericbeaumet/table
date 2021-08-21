package main

import (
	"github.com/jointable/table/cmd/api/routes"
	"github.com/jointable/table/pkg/service"
)

func main() {
	s := service.New()

	if err := routes.Up(s); err != nil {
		panic(err)
	}

	panic(s.Run("localhost:8080"))
}
