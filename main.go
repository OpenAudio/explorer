package main

import (
	"fmt"
	"sync"

	"github.com/AudiusProject/explorer/config"
	"github.com/AudiusProject/explorer/server"
)

func main() {
	config := config.NewConfig()

	s := server.New(config)

	var wg sync.WaitGroup

	wg.Go(func() {
		if err := s.Start(); err != nil {
			fmt.Println("oops", err)
		}
	})

	wg.Go(func() {
		// put indexers in here
		return
	})

	wg.Wait()
}
