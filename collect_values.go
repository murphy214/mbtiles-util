package mbutil

import (
	"github.com/murphy214/vector-tile-go"
	"fmt"
) 

// map query
func (mbtiles *Mbtiles) CollectValuesString(layer_names []string, field string) map[string]string {
	tiles := mbtiles.GetAllTiles()
	mymap := map[string]string{}
	for _,tile := range tiles {
		bs,err := mbtiles.Query(tile)
		if err != nil {
			fmt.Println(err)
		}
		tilev,err := vt.NewTile(bs)
		if err != nil {
			fmt.Println(err)
		}
		for _,layer_name := range layer_names {
			layer,boolval := tilev.LayerMap[layer_name]
			if boolval {
				for layer.Next() {
					feat,err := layer.Feature()
					if err != nil {
						fmt.Println(err)
					}
					myval,boolval := feat.Properties[field]
					if boolval {
						myvals,boolval := myval.(string)
						if boolval {
							mymap[myvals] = ""
						}
					}
				}
			}
		}
	}
	// mylist := make([]string,len(mymap))
	// pos := 0
	// for k := range mymap {
	// 	mylist[pos] = k 
	// 	pos++
	// }
	return mymap
}