package cmd

import (
	util "github.com/murphy214/mbtiles-util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

)

var outfilename string
var filterjson string

func init() {
	rootCmd.AddCommand(filterCmd)
	rootCmd.PersistentFlags().StringVarP(&outfilename, "outfilename", "o", "", "X/Y/Z TILE")
	viper.BindPFlag("outfilename", rootCmd.PersistentFlags().Lookup("outfilename"))
	rootCmd.PersistentFlags().StringVarP(&filterjson, "filterjson", "j", "", "X/Y/Z TILE")
	viper.BindPFlag("filterjson", rootCmd.PersistentFlags().Lookup("filterjson"))
}

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "filterjson",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		util.CreateMbtilesOut(filename,outfilename,filterjson)
		
	},
}
