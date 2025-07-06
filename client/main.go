package main

import (
	"GophKeeper/client/cmd"
	"fmt"
	"os"
)

func main() {

	// cfg := config.NewConfig()
	// cmd.Config = *cfg

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
