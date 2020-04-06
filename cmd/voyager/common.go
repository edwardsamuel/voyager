package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func serve(fns ...func(stopC <-chan struct{}) (<-chan error, error)) error {
	stopC := make(chan struct{})
	errC := make(chan error)

	for _, fn := range fns {
		ec, err := fn(stopC)
		if err != nil {
			return err
		}
		go func() {
			for er := range ec {
				errC <- er
			}
		}()
	}
	{
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		select {
		case s := <-sigs:
			log.Println("Got a signal", s)
		case e := <-errC:
			log.Println("Got an error", e)
		}
		close(stopC)
		log.Println("Completed shut down")
	}
	go func() {
		for er := range errC {
			log.Println(er)
		}
	}()
	return nil
}
