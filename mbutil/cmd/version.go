package cmd

import (
	"fmt"
	//"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)

}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hugo",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("testing fuck it")
		fmt.Println(filename)
	},
}
