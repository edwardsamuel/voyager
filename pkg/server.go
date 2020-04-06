package voyager

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

func ServeTCP(address string, stopC <-chan struct{}, handler func(conn net.Conn) error) (<-chan error, error) {
	listener, e := net.Listen("tcp", address)
	if e != nil {
		return nil, e
	}
	errC := make(chan error)
	go func() {
		for {
			select {
			case _, ok := <-stopC:
				if ok {
					break
				}
			default:
				conn, e := listener.Accept()
				if e != nil {
					log.Println("Exception accepting connection", e)
					errC <- e
				} else {
					go func() {
						defer func() {
							log.Println(conn.Close())
						}()
						errC <- handler(conn)
					}()
				}
			}
		}
	}()
	return errC, nil
}
func ServeHTTP(address string, stopC <-chan struct{}, handlers map[string]http.Handler) (<-chan error, error) {
	for p, h := range handlers {
		http.Handle(p, h)
	}
	listener, e := net.Listen("tcp", address)
	server := http.Server{}
	if e != nil {
		return nil, e
	}
	errC := make(chan error)
	go func() {
		errC <- server.Serve(listener)
		close(errC)
	}()
	go func() {
		<-stopC
		errC <- server.Shutdown(context.Background())
	}()

	return errC, nil

}
func ServeGRPC(address string, stopC <-chan struct{}, registerFn func(server *grpc.Server)) (<-chan error, error) {
	log.Println("Opening listener at ", address)
	listener, e := net.Listen("tcp", address)
	if e != nil {
		return nil, e
	}
	server := grpc.NewServer()
	registerFn(server)
	errC := make(chan error)
	go func() {
		log.Println("Ready to serve")
		errC <- server.Serve(listener)
		log.Println("Stopped serving")
		close(errC)
	}()

	go func() {
		<-stopC
		log.Println("Shutting down")
		server.GracefulStop()
	}()
	return errC, nil
}
