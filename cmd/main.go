package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/tombooth/home"
	"github.com/tombooth/home/config"
	"github.com/tombooth/home/web"
)

func main() {

	log.Println("Loading world from configuration")
	world, err := config.World()

	if err != nil {
		panic(err)
	}

	log.Println("Loading groups from configuration")
	groups, err := config.Groups(world)

	if err != nil {
		panic(err)
	}

	log.Println("Seeding the world")
	if err = world.Seed(); err != nil {
		panic(err)
	}

	stepTicker := time.NewTicker(config.StepInterval)
	transitions := make(chan home.WorldTransition, 100)

	go func() {
		for {
			select {

			case <- stepTicker.C:
				log.Println("Stepping the world forward")
				if err := world.Step(); err != nil {
					log.Println(err)
				}

			case transition := <- transitions:
				log.Printf("Applying a transition: %v\n", transition)
				if err := world.Apply(transition); err != nil {
					log.Println(err)
				}

			}
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Printf("Starting web server on %s\n", config.Port)

		http.ListenAndServe(
			config.Port,
			web.Routes(world, groups, transitions),
		)

		wg.Done()
	}()

	wg.Wait()

}
