package main

import (
	"fmt"
	"time"

	"github.com/schollz/peerdiscovery"
	_ "github.com/schollz/peerdiscovery"
)

func main() {
	discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{
		Limit: 1, Delay: time.Second * 20,
		AllowSelf: true,
	})
	for _, d := range discoveries {
		fmt.Printf("discovered '%s'\n", d.Address)
	}
}
