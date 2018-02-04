package mbutil

import (

	"github.com/paulmach/go.geojson"
	m "github.com/murphy214/mercantile"
	"github.com/murphy214/rdp"
)



type Simplify_Config struct {
	RDP_Bool bool // bool for an rdp implementation
	Limit bool // bool for a hard limit on the number of features in tiles
	Window_Bool bool // bool for the filter of a feature ocupying so much a dimmension of a tile
	Feature_Limit int // number of features limited per tile
	Window_Percent float64 // window percent 
	DeltaX float64 // the delta x of tile 
	DeltaY float64 // the deltay of a tile
	Zoom int
}


// a windowign bool
func (config *Simplify_Config) Window(featbds []float64) bool {
	// this makes sure the feature isnt a point!
	if featbds[2] != featbds[0] {
		if config.Window_Bool {
			return (((featbds[2] - featbds[0]) / config.DeltaX) > config.Window_Percent) || 
			(((featbds[3] - featbds[1]) / config.DeltaY) > config.Window_Percent)
		} else {
			return true
		}
	} else {
		return true
	}
}

func (config *Simplify_Config) RDP(feat *geojson.Feature) *geojson.Feature {
	if config.RDP_Bool {
		return rdp.RDP(feat,config.Zoom)
	} else {
		return rdp.RDP(feat,config.Zoom)
	}
}

// 
func New_Simplify_Config(rdpbool bool,limitbool bool,windowbool bool,tileid m.TileID) Simplify_Config {
	bds := m.Bounds(tileid)
	deltax := bds.E - bds.W
	deltay := bds.N - bds.S
	config := Simplify_Config{
		RDP_Bool:rdpbool,
		Limit:limitbool,
		Window_Bool:windowbool,
		Feature_Limit:200000,
		Window_Percent:1.0/256.0,
		DeltaX:deltax,
		DeltaY:deltay,
		Zoom:int(tileid.Z),
	}
	return config
}

func (config *Simplify_Config) Simplify(i *geojson.Feature) *geojson.Feature {
	if config.Window(i.Geometry.Get_BoundingBox()) {
		i = config.RDP(i)
		return i
	}
	return &geojson.Feature{}
}