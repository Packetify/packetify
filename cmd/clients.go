package cmd

import (
	"fmt"
	"github.com/Packetify/packetify/networkHandler"
	"github.com/spf13/cobra"
	"log"
)

var(
	virtualInterface string
	infoCommand = &cobra.Command{
        Use:   "info",
        Short: "Get information about the wifi clients",
        Long:  "Get information about the wifi clients",
        Run: func(cmd *cobra.Command, args []string) {
			if virtualInterface == "" {
				cmd.Help()
				return
			}
            wifi,err := networkHandler.NewWIFI(virtualInterface)
			if err != nil {
				log.Println(err)
				return
			}
			clientsInfo,err := wifi.IWClientsInfo()
			if err != nil {
                log.Println(err)
                return
            }
			fmt.Println(clientsInfo)
        },
    }
	clientsCommand = &cobra.Command{
		Use:   "clients",
        Short: "Manage clients",
        Long:  `Manage clients`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
                cmd.Help()
                return
            }
		},
	}
)

func init(){
	rootCmd.AddCommand(clientsCommand)
	clientsCommand.AddCommand(infoCommand)
	infoCommand.Flags().StringVarP(
		&virtualInterface,
		"virtualiface",
		"v",
		"",
		"The virtual interface thatpacketify use",
	)

}