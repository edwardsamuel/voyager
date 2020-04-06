package main

import (
	"github.com/spf13/cobra"
	voyager "github.com/vvarma/voyager/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net/http"
)

var (
	grpcAddress string
	httpAddress string
	tcpAddress  string
	serverCmd   = &cobra.Command{
		Use:  "server",
		Long: "Start the probe test server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(func(stopC <-chan struct{}) (errors <-chan error, e error) {
				return voyager.ServeGRPC(grpcAddress, stopC, func(server *grpc.Server) {
					npi := voyager.NewGrpcProbe()
					voyager.RegisterNetworkProbeServer(server, npi)
					reflection.Register(server)
				})
			}, func(stopC <-chan struct{}) (errors <-chan error, e error) {
				return voyager.ServeHTTP(httpAddress, stopC, map[string]http.Handler{
					"/voyager": voyager.NewHttpProbe(),
				})
			}, func(stopC <-chan struct{}) (errors <-chan error, e error) {
				return voyager.ServeTCP(tcpAddress, stopC, voyager.NewTCPProbe())
			})
		}}
)

func init() {
	serverCmd.PersistentFlags().StringVar(&grpcAddress, "address", "0.0.0.0:4891", "Address to start accepting grpc connections")
	serverCmd.PersistentFlags().StringVar(&httpAddress, "http-address", "0.0.0.0:4892", "Address to start accepting http connections")
	serverCmd.PersistentFlags().StringVar(&tcpAddress, "tcp-address", "0.0.0.0:4893", "Address to start accepting tcp connections")
	rootCmd.AddCommand(serverCmd)
}
