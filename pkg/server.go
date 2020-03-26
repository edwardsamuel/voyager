package voyager

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func ServeGRPC(address string, registerFn func(server *grpc.Server)) error {

	log.Println("Opening listener at ", address)
	listener, e := net.Listen("tcp", address)
	if e != nil {
		return e
	}
	server := grpc.NewServer()
	registerFn(server)
	stopC := make(chan struct{})
	errC := make(chan error)
	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()
		log.Println("Ready to serve")
		errC <- server.Serve(listener)
		log.Println("Stopped serving")
	}()
	go func() {
		for er := range errC {
			log.Println(er)
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		<-stopC
		log.Println("Shutting down")
		server.GracefulStop()
	}()
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
		wg.Wait()
		close(errC)
		log.Println("Completed shut down")
	}
	return nil
}
