package cmd

import (
	util "github.com/murphy214/mbtiles-util"
	"github.com/spf13/cobra"
	//"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(removeLayers)
	//rootCmd.PersistentFlags().StringVar(&tile, "tile", "0/0/0", "X/Y/Z TILE")
	//viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))
}

var removeLayers = &cobra.Command{
	Use:   "removelayers",
	Short: "Removes given layers and updates metadata",
	Long:  `Usage: 
mbutil removelayers -f pretty.mbtiles transportation transportation_name
	`,
	Run: func(cmd *cobra.Command, args []string) {
		layernames := args
		if len(filename) == 0 {
			if len(args) > 0 {
				filename = args[0]
				layernames = args[1:]
			}
		}
		util.RemoveLayers(filename,layernames)	
	},
}
