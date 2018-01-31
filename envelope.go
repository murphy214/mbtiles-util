package mbutil

import (
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	//"fmt"
	"sync"
	fm "github.com/murphy214/feature-map"
	"fmt"
)

// recursively drills until the max zoom is reached
func (mbutil Mbtiles) Make_Zoom_Drill(k m.TileID, v []*geojson.Feature,endsize int) {

	outputsize := int(k.Z) + 1
	cc := make(chan map[m.TileID][]*geojson.Feature)
	for _, i := range v {
		go func(k m.TileID, i *geojson.Feature, cc chan map[m.TileID][]*geojson.Feature) {
			if i.Geometry == nil {
				cc <- map[m.TileID][]*geojson.Feature{}
			} else if i.Geometry.Type == "Polygon" {
				partmap := fm.Children_Polygon(i, k) 
				cc <- partmap
			} else if i.Geometry.Type == "LineString" {
				partmap := fm.Env_Line(i, int(k.Z+1))
				partmap = fm.Lint_Children_Lines(partmap, k)
				cc <- partmap
			} else if i.Geometry.Type == "Point" {
				partmap := map[m.TileID][]*geojson.Feature{}
				pt := i.Geometry.Point
				tileid := m.Tile(pt[0], pt[1], int(k.Z+1))
				partmap[tileid] = append(partmap[tileid], i)
				cc <- partmap
			} else {
				fmt.Println("SHIT")
			}
		}(k, i, cc)
	}

	// collecting all into child map
	childmap := map[m.TileID][]*geojson.Feature{}
	for range v {
		partmap := <-cc
		for kk, vv := range partmap {
			if len(vv) > 0 {
				childmap[kk] = append(childmap[kk], vv...)
			}
		}
	}
	// iterating through each value in the child map and waiting to complete
	//var wg sync.WaitGroup
	var wg sync.WaitGroup
	for kkk, vvv := range childmap {
		//childmap = map[m.TileID][]*geojson.Feature{}
		wg.Add(1)
		go func(kkk m.TileID, vvv []*geojson.Feature) {
			mbutil.Make_Tile_Geojson(kkk, vvv)
			wg.Done()

		}(kkk, vvv)
	}
	wg.Wait()
	
	//wg.Wait()
	if endsize != outputsize && len(childmap) != 0 {
		var wgg sync.WaitGroup
		for kkk, vvv := range childmap {
			wgg.Add(1)
			go func(kkk m.TileID, vvv []*geojson.Feature) {
				mbutil.Make_Zoom_Drill(kkk,vvv,endsize)
				wgg.Done()
			}(kkk,vvv)
		}
		wgg.Wait()

	} else {
	}
}

