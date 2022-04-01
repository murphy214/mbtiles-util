package mbutil

import (
	"github.com/murphy214/pbf"
	"os"
	m "github.com/murphy214/mercantile"
	"github.com/murphy214/vector-tile-go"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"errors"
)

type Filter struct {
	Key        string
	FilterType string `json:"filter_type"`
	Operator   string `json:"operator"`
	Value      interface{}
	Filters    []*Filter
}

func Operate(operator string, reference, compare interface{}) bool {
	switch operator {
	case "==":
		return reference == compare
	case "!=":
		return reference != compare
	case "<=":
		return reference.(float64) <= compare.(float64)
	case ">=":
		return reference.(float64) >= compare.(float64)
	}
	return false
}

func (filt *Filter) Filter(properties map[string]interface{}) bool {
	switch filt.FilterType {
	case "string":
		myval := properties[filt.Key]
		mystr, boolval := filt.Value.(string)
		mystr2, boolval2 := myval.(string)
		if boolval && boolval2 {
			return Operate(filt.Operator, mystr, mystr2)
		}
		return false
	case "float":
		myval := properties[filt.Key]
		mystr, boolval := filt.Value.(float64)
		mystr2, boolval2 := myval.(float64)
		if boolval && boolval2 {
			return Operate(filt.Operator, mystr, mystr2)
		}
		return false
	case "all":
		boolval := true
		for _, f := range filt.Filters {
			boolval = boolval && f.Filter(properties)
		}
		return boolval
	case "any":
		boolval := false
		for _, f := range filt.Filters {
			boolval = boolval || f.Filter(properties)
		}

		return boolval

	}
	return false
}

type FilterLayer struct {
	LayerNameSource  string
	LayerName        string
	MinZoom, MaxZoom int
	Filter           *Filter
}

func (filterlayer *FilterLayer) ApplyFilter(properties map[string]interface{}, layername string, zoom int) bool {
	layernamebool := layername == filterlayer.LayerNameSource
	zoombool := (zoom <= filterlayer.MaxZoom) && (zoom >= filterlayer.MinZoom)
	if layernamebool {
		return layernamebool && zoombool && filterlayer.Filter.Filter(properties)
	} else {
		return true
	}
}

//
type FilterTotal struct {
	FilterLayers []*FilterLayer `json:"filter_layers"`
	FilterMap    map[string][]*FilterLayer
}

func (filtertotal *FilterTotal) Init() {
	filtertotal.FilterMap = map[string][]*FilterLayer{}
	for _, filterlayer := range filtertotal.FilterLayers {
		val, boolval := filtertotal.FilterMap[filterlayer.LayerNameSource]
		if !boolval {
			val = []*FilterLayer{}
		}
		val = append(val, filterlayer)
		filtertotal.FilterMap[filterlayer.LayerNameSource] = val
	}
}

//
func (filtertotal *FilterTotal) Filter(properties map[string]interface{}, layername string, zoom int) bool {
	val, boolval := filtertotal.FilterMap[layername]
	if !boolval {
		return true
	} else if boolval {
		for _, i := range val {
			boolval2 := i.ApplyFilter(properties, layername, zoom)
			if boolval2 {
				return true
			}
		}
	}
	return false
}

func (filtertotal *FilterTotal) SkipLayer(layername string) bool {
	val,boolval := filtertotal.FilterMap[layername]
	if boolval {
		return len(val) == 0 
	}
	return true
}



// marshals the json
func (filtertotal *FilterTotal) Marshall() ([]byte, error) {
	filtertotal.FilterMap = map[string][]*FilterLayer{}
	bs, err := json.Marshal(filtertotal)
	return bs, err
}

// reads a given filter
func ReadFilter(bs []byte) (*FilterTotal, error) {
	var mm FilterTotal
	err := json.Unmarshal(bs, &mm)
	if err != nil {
		return &FilterTotal{}, err
	}
	mm.Init()
	return &mm, nil
}


// applies a filter on an entire vector tile array
func (filtertotal *FilterTotal) FilterVectorTile(bs []byte,tileid m.TileID) ([]byte) {
	tile,err := vt.NewTile(bs)
	if err != nil {
		fmt.Println(err)
	}

	totalbs := []byte{}
	for _,layer := range tile.LayerMap {
		if filtertotal.SkipLayer(layer.Name) {
			//fmt.Println(layer.Name)
			tmpbs := layer.Buf.Pbf[layer.StartPos:layer.EndPos]
			beg := append([]byte{26}, pbf.EncodeVarint(uint64(len(tmpbs)))...)
			tmpbs = append(beg, tmpbs...)
			//fmt.Println(len(tmpbs))
			totalbs = append(totalbs,tmpbs...)
		} else {
			config := vt.Config{
				TileID:tileid,
				Name:layer.Name,
				Extent:int32(layer.Extent),
				Version:layer.Version,
			}
			layerwrite := vt.NewLayerConfig(config)


			for layer.Next() {
				feature,err := layer.Feature()
				if err != nil {
					fmt.Println(err,layer.Name)
				}
				filterbool := filtertotal.Filter(feature.Properties,layerwrite.Name,int(tileid.Z))
				if filterbool {
					geom,err := feature.LoadGeometryRaw()
					if err != nil {
						fmt.Println(err,layer.Name)
					} else {
						layerwrite.AddFeatureRaw(0,feature.Geom_int,geom,feature.Properties)
					}

				}
			}
			tmpbs := layerwrite.Flush()
			totalbs = append(totalbs,tmpbs...)
		}
	}
	return totalbs
}



func CreateMbtilesOut(inmbtiles,outmbtiles string,infilterjson string) {
	fbs,err := ioutil.ReadFile(infilterjson)
	if err != nil {
		fmt.Println(err)
	}
	filtertotal,err := ReadFilter(fbs)
	if err != nil {
		fmt.Println(err)
	}
	confige := Config{LayerName: "Test",
		FileName:        outmbtiles,
		LayerProperties: map[string]interface{}{"shit": 10},
		MinZoom:         0,
		MaxZoom:14,
	}
	os.Remove(outmbtiles)
	mbtile,err := CreateDB(confige)
	if err != nil {
		fmt.Println(err)
	}

	oldmbtiles,_ := ReadMbtiles(inmbtiles)
	tiles := oldmbtiles.GetAllTiles()
	for pos,tile := range tiles {
		bs,err := oldmbtiles.Query(tile)
		if err != nil {
			fmt.Println(err)
		}
		bs = filtertotal.FilterVectorTile(bs,tile)
		mbtile.AddTile(tile,bs)
		if pos%1000==0 {
			fmt.Printf("\r[%d/%d]",pos,len(tiles))
		}
	}
	fmt.Println()
	mbtile.Commit()
	UpdateMetaDataJSON(outmbtiles)
}

// filters a geometry based on size
func FilterSizeGeom(tileid m.TileID,bs []byte,layername string,size float64,maxzoom int) ([]byte,error) {
	tile, err := vt.NewTile(bs)
	if err != nil {
		fmt.Println(err)
	}
	if uint64(maxzoom) < tileid.Z {
		return bs,nil
	}
	i := 0
	total := 0
	layer := tile.LayerMap["Test"]
	//layerwrite := vt.NewLayer(tileid,layername)

	layernew := *layer
	layerwrite,err := layernew.ToLayerWrite(tileid)
	if err != nil {
		return []byte{},errors.New("Layerwrite not initalized!")
	}
	layerwrite.Values_Bytes = []byte{}
	layerwrite.Values_Map = map[interface{}]uint32{}
	layerwrite.Keys_Bytes = []byte{}
	layerwrite.Keys_Map = map[string]uint32{}
	layerwrite.Features = []byte{}


	for layer.Next() {
		feat,err := layer.Feature()
		if err != nil {
			fmt.Println(err)
		}
		geom,err := feat.LoadGeometry()
		if err != nil {
			fmt.Println(err)
		}
		bbox := vt.Get_BoundingBox(geom)
		if len(bbox) >= 4 {
			west, south, east, north := bbox[0],bbox[1],bbox[2],bbox[3]
			dist := math.Sqrt(math.Pow(east-west,2)+math.Pow(north-south,2))
			if (dist>size) {
				//fmt.Println(west,south,east,north,dist)
				//geoms,_ := feat.LoadGeometryRaw()
				//featg,_ := feat.ToGeoJSON(tileid)

				//layerwrite.AddFeatureRaw(feat.ID,feat.Geom_int,geoms,feat.Properties)

				layerwrite.AddFeatureLazy(feat)
				//layerwrite.AddFeature(featg)
				i++
			}
		} else {
			fmt.Println("bbox failed")
		}
		total++
	}
	//fmt.Printf("Zoom: %d,Features added: %d, features before: %d\n",tileid.Z,i,total)
	if i > 0 {
		return layerwrite.Flush(),err 
	} else {
		return []byte{},nil
	}
}


func RemoveDenseMbtiles(mbfilename string,outfn string,geomsize,maxzoom int) {
	inmbtiles,_ := ReadMbtiles(mbfilename)
	confige := Config{LayerName: "Test",
	FileName:        outfn,
	LayerProperties: map[string]interface{}{"shit": 10},
	MinZoom:         0,
	MaxZoom:14,
	}
	os.Remove(outfn)
	mbtile,_ := CreateDB(confige)
	tiles := inmbtiles.GetAllTiles()

	for pos,tile := range tiles {
		bs,err := inmbtiles.Query(tile)
		bss,err := FilterSizeGeom(tile,bs,"Test",float64(geomsize),maxzoom)
		if err != nil {
			fmt.Println(err)
		}
		if len(bss)>0 {
			mbtile.AddTile(tile,bss)
		}
		if pos%100 == 0 {
			fmt.Printf("%d [%d/%d]\n",tile.Z,pos,len(tiles))
		}
	}
	mbtile.Commit()
}