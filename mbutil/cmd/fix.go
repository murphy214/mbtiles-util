package cmd

import (
	util "github.com/murphy214/mbtiles-util"
	"github.com/spf13/cobra"
	//"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(fixMeta)
	//rootCmd.PersistentFlags().StringVar(&tile, "tile", "0/0/0", "X/Y/Z TILE")
	//viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))
}

var fixMeta = &cobra.Command{
	Use:   "fix",
	Short: "Fixes the view table to table.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(filename) == 0 {
			if len(args) > 0 {
				filename = args[0]
			}
		}
		util.FixViewToTable(filename)	
	},
}
