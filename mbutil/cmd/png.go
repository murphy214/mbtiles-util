package cmd

import (
	util "github.com/murphy214/mbtiles-util"
	"github.com/murphy214/mbtiles-util/draw"
	m "github.com/murphy214/mercantile"
	"github.com/murphy214/vector-tile-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"fmt"
)

//var tile string

//var resolution int
var out string

func init() {
	rootCmd.AddCommand(pngCmd)
	//rootCmd.PersistentFlags().StringVar(&tile, "tile", "", "X/Y/Z TILE")
	viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))
	//rootCmd.PersistentFlags().IntVar(&resolution, "res", 4096, "Resolution of tile")
	viper.BindPFlag("res", rootCmd.PersistentFlags().Lookup("res"))
	rootCmd.PersistentFlags().StringVar(&out, "out", "output.png", "Output PNG filename")
	viper.BindPFlag("out", rootCmd.PersistentFlags().Lookup("out"))
}

var pngCmd = &cobra.Command{
	Use:   "png",
	Short: "Draws a to a png at a given resolution",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		mbtile, _ := util.ReadMbtiles(filename)
		var tileid m.TileID
		if tile == "" {
			tileid = mbtile.SingleTile()
		} else {
			tileid = m.Strtile(tile)
		}
		if resolution == 100 {
			resolution = 4096
		}
		fmt.Printf("Drawing %+v tile\n", tileid)
		bytevals, _ := mbtile.Query(tileid)
		features, _ := vt.ReadTile(bytevals, tileid)
		drawer.WriteFeaturesPNG(features, tileid, resolution, out)
	},
}
