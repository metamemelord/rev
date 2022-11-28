package commands

import (
	"github.com/metamemelord/rev/connect"
	"github.com/metamemelord/rev/server"
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:     "rev",
	Short:   "Rev is a reverse proxy",
	Version: "Rev 0.0.1",
}

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Run server for clients to connect",
	Run: func(cmd *cobra.Command, args []string) {
		server.Init(cmd.Flags().Lookup("port").Value.String())
	},
	Version: rootCommand.Version,
}

var connectCommand = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a running Rev server to start a tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		connect.Init(cmd.Flags().Lookup("server-host").Value.String(),
			cmd.Flags().Lookup("server-port").Value.String(),
			cmd.Flags().Lookup("service-user").Value.String(),
			cmd.Flags().Lookup("service-name").Value.String(),
			cmd.Flags().Lookup("destination-protocol").Value.String(),
			cmd.Flags().Lookup("destination-host").Value.String(),
			cmd.Flags().Lookup("destination-port").Value.String())
	},
	Version: rootCommand.Version,
}

func init() {
	rootCommand.AddCommand(serverCommand, connectCommand)
	rootCommand.CompletionOptions.DisableDefaultCmd = true
	rootCommand.SetVersionTemplate("")
	serverCommand.Flags().StringP("port", "p", "8080", "Port to run the server")

	connectCommand.Flags().StringP("server-host", "s", "", "Server hostname or IP")
	connectCommand.Flags().StringP("server-port", "p", "", "Server port")
	connectCommand.Flags().StringP("destination-protocol", "", "http", "Destination service protocol")
	connectCommand.Flags().StringP("destination-host", "", "localhost", "Destination service hostname")
	connectCommand.Flags().StringP("destination-port", "d", "", "Destination service port")
	connectCommand.Flags().StringP("service-user", "u", "", "Service user")
	connectCommand.Flags().StringP("service-name", "n", "", "Service name")

	connectCommand.MarkFlagRequired("server-host")
	connectCommand.MarkFlagRequired("server-port")
	connectCommand.MarkFlagRequired("destination-port")
	connectCommand.MarkFlagRequired("service-user")
	connectCommand.MarkFlagRequired("service-name")
}

func Run() {
	rootCommand.Execute()
}
