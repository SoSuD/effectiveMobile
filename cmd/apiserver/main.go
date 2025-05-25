package main

import (
	"effectiveMobile/internal/app/apiserver"
	"fmt"
)

func main() {
	config := apiserver.NewConfig()
	fmt.Println(config)
	if err := apiserver.Start(config); err != nil {
		panic(err)
	}
}
