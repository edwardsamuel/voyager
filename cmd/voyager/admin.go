package main

import (
	"github.com/spf13/cobra"
	voyager "github.com/vvarma/voyager/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	adminAddress string
	adminCmd     = &cobra.Command{
		Long: "Start the admin server",
		Use:  "admin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(func(stopC <-chan struct{}) (errors <-chan error, e error) {
				return voyager.ServeGRPC(adminAddress, stopC, func(server *grpc.Server) {
					a := voyager.NewAdminImpl()
					voyager.RegisterAdminServer(server, a)
					reflection.Register(server)
				})
			})
		}}
)

func init() {
	adminCmd.PersistentFlags().StringVar(&adminAddress, "address", "0.0.0.0:14891", "Address to start accepting grpc connections")
	rootCmd.AddCommand(adminCmd)
}
