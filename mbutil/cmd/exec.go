package cmd

import (
	"strings"

	util "github.com/murphy214/mbtiles-util"
	m "github.com/murphy214/mercantile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"fmt"
)

var gofile string
func init() {
	rootCmd.AddCommand(execCmd)
	viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))
	rootCmd.PersistentFlags().StringVarP(&gofile, "gofile", "g", "mine.go", "Go file to run on tile bytes")
	viper.BindPFlag("gofile", rootCmd.PersistentFlags().Lookup("gofile"))
	rootCmd.PersistentFlags().Bool("template", false, "creates template go file")
	viper.BindPFlag("myVipe", rootCmd.PersistentFlags().Lookup("template"))


}
var mytml = `
package main

import (
	"github.com/murphy214/vector-tile-go"
	mbutil "github.com/murphy214/mbtiles-util"
	"fmt"
)

func main() {
	bs,_ := mbutil.ReadStdin()
	tile, _ := vt.NewTile(bs)
	for k := range tile.LayerMap {
		fmt.Println(k)
	}
}
`
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Executes a go script on a give tile.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		boolval := viper.GetBool("myVipe")
		if boolval&&len(outfilename)>0&&strings.HasSuffix(outfilename,".go") {
			ioutil.WriteFile(outfilename,[]byte(mytml),0677)
		} else {
			mbtile, _ := util.ReadMbtiles(filename)
			var tileid m.TileID
			if tile == "" {
				tileid = mbtile.SingleTile()
			} else {
				tileid = m.Strtile(tile)
			}
	
			fmt.Printf("Executing %+v tile %s go file on %s mbtiles file.\n", tileid,gofile,filename)
			bytevals, _ := mbtile.Query(tileid)
			util.ExecCmdBytes([]string{"go","run",gofile},bytevals)
			//features, _ := vt.ReadTile(bytevals, tileid)
			//screen := drawer.NewGrid(tileid, resolution)
			//screen.PaintFeatures(features)
			//screen.Screen.Draw()
		}


	},
}
