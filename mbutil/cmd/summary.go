package cmd

import (
	"fmt"
	util "github.com/murphy214/mbtiles-util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"os"
	"sort"
)

func CreateMetaData(filename string) {
	mbtile, _ := util.ReadMbtiles(filename)
	rows, err := mbtile.Tx.Query("SELECT name,value from metadata;")
	if err != nil {
		fmt.Println(err)
	}
	metadata := [][]string{}
	var metajson string
	for rows.Next() {
		var val1, val2 string
		rows.Scan(&val1, &val2)
		if val1 != "json" {
			metadata = append(metadata, []string{val1, val2})
		} else {
			metajson = val2
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "VALUE"})
	table.AppendBulk(metadata)
	table.Render()
	fmt.Println("Layer Metadata:", metajson)

}

// creates a padding for whitespace
func WhiteSpaceMake(mystring string, size int) string {
	paddingsize := size - len(mystring)
	padding := ""
	for i := 0; i < paddingsize; i++ {
		padding += " "
	}
	return padding + mystring
}

/*
Prints generalized metadata to a table.
*/
func CreateTileStats(filename string) {
	mbtile, _ := util.ReadMbtiles(filename)
	totalmap := map[int][2]int{}
	for mbtile.Old_Total != mbtile.Total {
		metadata := mbtile.ChunkTilesMeta(1000)

		// iterating through each metadata tile
		for k, v := range metadata {
			total, boolval := totalmap[int(k.Z)]
			if !boolval {
				total = [2]int{0, 0}
			}
			total[0] += 1
			total[1] += v
			totalmap[int(k.Z)] = total
		}
	}
	mbsize := math.Pow(1000, 2)
	totalfloat := map[int][2]float64{}
	for k, v := range totalmap {
		totalfloat[k] = [2]float64{float64(v[0]), float64(v[1]) / mbsize}
	}

	keys := []int{}
	for k := range totalmap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	data := [][]string{}
	totalcount := 0
	totalf := 0.0
	for _, k := range keys {
		v := totalfloat[k]

		row := []string{fmt.Sprintf("%d", k), fmt.Sprintf("%d", int(v[0])), fmt.Sprintf("%f", v[1])}
		data = append(data, row)
		totalcount += int(v[0])
		totalf += v[1]
	}

	footer := []string{"TOTAL", fmt.Sprintf("%d", totalcount), fmt.Sprintf("%f", totalf)}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ZOOM", "COUNT", "SIZE (MB)"})
	table.AppendBulk(data)
	table.SetFooter(footer)
	table.Render()

}

func init() {
	rootCmd.AddCommand(summaryCmd)
	viper.BindPFlag("filename", rootCmd.PersistentFlags().Lookup("filename"))
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Prints Summary Statistics in a table.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		CreateMetaData(filename)
		CreateTileStats(filename)
	},
}
