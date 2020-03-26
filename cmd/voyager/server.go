package main

import (
	"github.com/spf13/cobra"
	voyager "github.com/vvarma/voyager/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	address   string
	serverCmd = &cobra.Command{
		Use:  "server",
		Long: "Start the probe test server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return voyager.ServeGRPC(address, func(server *grpc.Server) {
				npi := voyager.NewGrpcProbe()
				voyager.RegisterNetworkProbeServer(server, npi)
				reflection.Register(server)
			})
		}}
)

func init() {
	serverCmd.PersistentFlags().StringVar(&address, "address", "0.0.0.0:4891", "Address to start accepting grpc connections")
	rootCmd.AddCommand(serverCmd)
}
