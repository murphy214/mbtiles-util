package cmd

import (
	"fmt"
	util "github.com/murphy214/mbtiles-util"
	m "github.com/murphy214/mercantile"
	"github.com/murphy214/vector-tile-go"
	"github.com/paulmach/go.geojson"
	"github.com/spf13/cobra"
	//"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(geojsonCmd)
	//rootCmd.PersistentFlags().StringVar(&tile, "tile", "0/0/0", "X/Y/Z TILE")
	//viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))
}

var geojsonCmd = &cobra.Command{
	Use:   "geojson",
	Short: "Prints a geojson string for a given tile.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		mbtile := util.ReadMbtiles(filename)
		var tileid m.TileID
		if tile == "" {
			tileid = mbtile.SingleTile()
		} else {
			tileid = m.Strtile(tile)
		}

		fc := &geojson.FeatureCollection{Features: vt.ReadTile(mbtile.Query(tileid), tileid)}
		bytevals, err := fc.MarshalJSON()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(bytevals))
	},
}
