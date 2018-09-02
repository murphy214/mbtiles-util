package cmd

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	util "github.com/murphy214/mbtiles-util"
	"github.com/murphy214/mbtiles-util/mbutil/vector-tile/2.1"
	m "github.com/murphy214/mercantile"
	"github.com/spf13/cobra"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	//"github.com/spf13/viper"
)

func Server(gvt util.Mbtiles, port int) {
	// creating mutex as this process does need to block as its drilling tiles.
	var mm sync.Mutex
	srv := &http.Server{Addr: fmt.Sprintf(":%s", strconv.Itoa(port))}
	test := "Test"
	tile := &vector_tile.Tile{Layers: []*vector_tile.Tile_Layer{{Name: &test}}}
	bs, _ := proto.Marshal(tile)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mm.Lock()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		vals := strings.Split(r.URL.Path, "/")

		if len(vals) == 4 {
			// getting x,y,z tile
			z, x, y := vals[1], vals[2], vals[3]
			znew, _ := strconv.ParseInt(z, 10, 64)
			xnew, _ := strconv.ParseInt(x, 10, 64)
			ynew, _ := strconv.ParseInt(y, 10, 64)
			//ynew = int64((1 << uint64(znew)) - 1 - ynew)
			s := time.Now()
			data, _ := gvt.Query(m.TileID{xnew, ynew, uint64(znew)})
			fmt.Fprintf(w, "%s", string(data))
			fmt.Printf("Tile: %s | Time: %v\n", x+"/"+y+"/"+z, time.Now().Sub(s))

		} else {
			fmt.Fprintf(w, "%s", string(bs))

		}
		mm.Unlock()
	})
	srv.ListenAndServe()
}

var port int

func init() {
	rootCmd.AddCommand(servercmd)
	//viper.BindPFlag("tile", rootCmd.PersistentFlags().Lookup("tile"))
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "port for the tiles to be serverd on")
}

var servercmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a dead simple server with the mbtiles file given.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		mbtile, _ := util.ReadMbtiles(filename)

		Server(mbtile, port)

	},
}
