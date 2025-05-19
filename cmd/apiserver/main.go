package main

import "effectiveMobile/internal/app/apiserver"

func main() {
	config := apiserver.NewConfig()
	if err := apiserver.Start(config); err != nil {
		panic(err)
	}
}
