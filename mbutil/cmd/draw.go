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

var tile string
var resolution int

func init() {
	rootCmd.AddCommand(drawCmd)
	rootCmd.PersistentFlags().StringVar(&tile, "tile", "", "X/Y/Z TILE")
	viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))
	rootCmd.PersistentFlags().IntVar(&resolution, "res", 100, "Resolution of tile")
	viper.BindPFlag("res", rootCmd.PersistentFlags().Lookup("res"))
}

var drawCmd = &cobra.Command{
	Use:   "draw",
	Short: "Draws a tile at a given resolution",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		mbtile := util.ReadMbtiles(filename)
		var tileid m.TileID
		if tile == "" {
			tileid = mbtile.SingleTile()
		} else {
			tileid = m.Strtile(tile)
		}
		fmt.Printf("Drawing %+v tile\n", tileid)
		features := vt.ReadTile(mbtile.Query(tileid), tileid)
		screen := drawer.NewGrid(tileid, resolution)
		screen.PaintFeatures(features)
		screen.Screen.Draw()
	},
}
