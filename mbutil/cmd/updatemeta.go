package cmd

import (
	util "github.com/murphy214/mbtiles-util"
	"github.com/spf13/cobra"
	//"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(updateMeta)
	//rootCmd.PersistentFlags().StringVar(&tile, "tile", "0/0/0", "X/Y/Z TILE")
	//viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))

}

var updateMeta = &cobra.Command{
	Use:   "updatemeta",
	Short: "Updates underlying metadata of a tile.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(filename) == 0 {
			if len(args) > 0 {
				filename = args[0]
			}
		}

		if len(args) == 1 && len(keyvalue) == 1 {
			keyvalue = append(keyvalue,args[0])
		}

		if len(keyvalue) == 2 {
			util.UpdateMetaDataKV(filename,keyvalue[0],keyvalue[1])
		} else {
			util.UpdateMetaDataJSON(filename)	
		}

	},
}
