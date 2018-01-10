package mbutil

import (
	"fmt"
	m "github.com/murphy214/mercantile"
	"math"
	"github.com/paulmach/go.geojson"
	"github.com/murphy214/gotile-geobuf/vector-tile/2.1"
	"github.com/golang/protobuf/proto"
)

// atructure converting points to long lat respectively
type Point_Convert struct {
	DeltaX float64
	DeltaY float64
	Bds m.Extrema

}

// converts an xy coordinate into a point
func (converter *Point_Convert) Convert_XY(xy []int) []float64 {
	merc_point := []float64{float64(xy[0]) / 4096.0 * converter.DeltaX + converter.Bds.W,(4096.0 - float64(xy[1])) / 4096.0 * converter.DeltaY + converter.Bds.S}
	return Convert_Merc_Point(merc_point)
}

// brute force converts all points in an int alignment
func (converter *Point_Convert) Convert_Coords(coords [][][]int) [][][]float64 {
	total := [][][]float64{}
	for _,cont := range coords {
		newline := [][]float64{}	
		for _,pt := range cont {
			newline = append(newline,converter.Convert_XY(pt))
		}
		total = append(total,newline)
	}
	return total
}

const mercatorPole = 20037508.34

func Convert_Point_Merc(point []float64) []float64 {
	x := mercatorPole / 180.0 * point[0]

	y := math.Log(math.Tan((90.0+point[1])*math.Pi/360.0)) / math.Pi * mercatorPole
	y = math.Max(-mercatorPole, math.Min(y, mercatorPole))
	return []float64{x,y}
}

// converting points to long,lat
func Convert_Merc_Point(point []float64) []float64 {
	x := float64(point[0]) / (math.Pi/180.0) / 6378137.0
	y := 180.0/math.Pi*(2.0*math.Atan(math.Exp((float64(point[1])/6378137.0)))-math.Pi/2.0)
	return []float64{x,y}
}


func New_Point_Convert(tileid m.TileID) Point_Convert {
	bds := Create_Mercator_Bounds(tileid)
	return Point_Convert{Bds:bds,DeltaX:bds.E-bds.W,DeltaY:bds.N-bds.S}
}


// creates mercator bounds
func Create_Mercator_Bounds(tileid m.TileID) m.Extrema {
	bounds := m.Bounds(tileid)
	en := []float64{bounds.E,bounds.N} // east, north point
	ws := []float64{bounds.W,bounds.S} // west, south point

	// converting these
	en = Convert_Point_Merc(en)
	ws = Convert_Point_Merc(ws)

	// gettting north east west south
	east := en[0]
	north := en[1]
	west := ws[0]
	south := ws[1]
	bounds = m.Extrema{N:north,E:east,S:south,W:west}
	return bounds
}

// decodes a given delta
func DecodeDelta(nume uint32) int {
	num := int(nume)
	if (num % 2 == 1) {
		return (num + 1) / -2
	} else {
		return num / 2
	}
}
		
func Get_Command_Length(cmdLen uint32) (int32,int32) {
	cmd := cmdLen & 0x7;
    length := cmdLen >> 3;

    return int32(cmd),int32(length)
}

// decodes a given simple geometry
// returns it as an arbitary set of given lines
func DecodeGeometry(geom []uint32) [][][]int {
	pos := 0
	firstpt,currentpt := []int{},[]int{}
	newline := [][]int{}
	lines := [][][]int{}
	for pos < len(geom) {
		geomval := geom[pos]

		cmd,length := Get_Command_Length(geomval)

		// conde for a move to cmd
		if cmd == 1 {
			xdelta := DecodeDelta(geom[pos+1])
			ydelta := DecodeDelta(geom[pos+2])
			firstpt = []int{xdelta,ydelta}
			currentpt = firstpt
			//fmt.Println(firstpt)
			pos += 2

			if pos == len(geom) - 1 {
				lines = append(lines,[][]int{currentpt})
			}
		} else if cmd == 2 {
			newline = [][]int{firstpt}
			currentpos := pos + 1
			endpos := currentpos + int(length * 2)
			for currentpos < endpos {
				xdelta := DecodeDelta(geom[currentpos])
				ydelta := DecodeDelta(geom[currentpos+1])
				currentpt = []int{currentpt[0] + xdelta,currentpt[1] + ydelta}
				newline = append(newline,currentpt)
				currentpos += 2
			}

			pos = currentpos - 1
			lines = append(lines,newline)

		} else if cmd == 7 {
			newline := lines[len(lines)-1]
			newline = append(newline,newline[0])
			lines[len(lines)-1] = newline
		}

		//fmt.Println(cmd,length)
		pos += 1
	}
	return lines
}

// gets a vector tile value
func Get_Value(value *vector_tile.Tile_Value) interface{} {
	if value.StringValue != nil {
		return *value.StringValue
	} else if value.FloatValue != nil {
		return *value.FloatValue
	} else if value.DoubleValue != nil {
		return *value.DoubleValue
	} else if value.IntValue != nil {
		return *value.IntValue
	} else if value.UintValue != nil {
		return *value.UintValue
	} else if value.SintValue != nil {
		return *value.SintValue
	} else if value.BoolValue != nil {
		return *value.BoolValue
	} else {
		return ""
	} 
	return ""
}

// make keymap and values map 
func Make_Key_Value_Map(values []*vector_tile.Tile_Value,keys []string) (map[int]interface{},map[int]string) {
	// making value map
	valuemap := map[int]interface{}{}
	for i,value := range values {
		valuemap[i] = Get_Value(value)
	} 

	// making keymap
	keymap := map[int]string{}
	for i,key := range keys {
		keymap[i] = key
	}

	return valuemap,keymap
}


// returns a list of features associated with each vector tile
func Convert_Vt_Bytes(bytevals []byte,tileid m.TileID) []*geojson.Feature {
	tile := &vector_tile.Tile{}
	if err := proto.Unmarshal(bytevals, tile); err != nil {
		fmt.Println(err)
	}	

	converter := New_Point_Convert(tileid)

	newfeats := []*geojson.Feature{}
	for _,layer := range tile.Layers {
		// getting value and key map
		valuemap,keymap := Make_Key_Value_Map(layer.Values,layer.Keys)
		
		// making channel
		c := make(chan *geojson.Feature)

		//converting coordinates
		for _,feat := range layer.Features {
			go func(feat *vector_tile.Tile_Feature, c chan *geojson.Feature) {
				var geom *geojson.Geometry
				coords := converter.Convert_Coords(DecodeGeometry(feat.Geometry))
				if *feat.Type == vector_tile.Tile_POLYGON {
					geom = &geojson.Geometry{Polygon:coords,Type:"Polygon"}
				} else if *feat.Type == vector_tile.Tile_LINESTRING {
					geom = &geojson.Geometry{LineString:coords[0],Type:"LineString"}
				} else if *feat.Type == vector_tile.Tile_POINT {
					geom = &geojson.Geometry{Point:coords[0][0],Type:"Point"}
				} 

				properties := map[string]interface{}{}
				count := 0
				for count < len(feat.Tags) {
					properties[keymap[count]] = valuemap[count+1]
					count += 2
				}

				c <- &geojson.Feature{ID:feat.Id,Geometry:geom,Properties:properties}
			}(feat,c)
		}

		for range layer.Features {
			newfeats = append(newfeats,<-c)
		}


	}
	return newfeats
}