package mbutil

import (
	"fmt"
	"math"

	"github.com/golang/protobuf/proto"
	"github.com/murphy214/mbtiles-util/vector-tile/2.1"
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	"compress/gzip"
	"bytes"
	"io"
)

func GUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)
	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}
	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}
	resData = resB.Bytes()
	return
}

// atructure converting points to long lat respectively
type Point_Convert struct {
	DeltaX float64
	DeltaY float64
	Bds    m.Extrema
}

const mercatorPole = 20037508.34

// converts an xy coordinate into a point
func (converter *Point_Convert) Convert_XY(xy []int) []float64 {
	merc_point := []float64{float64(xy[0])/4096.0*converter.DeltaX + converter.Bds.W, (4096.0-float64(xy[1]))/4096.0*converter.DeltaY + converter.Bds.S}
	return Convert_Merc_Point(merc_point)
}

// brute force converts all points in an int alignment
func (converter *Point_Convert) Convert_Coords(coords [][][][]int) [][][][]float64 {
	total := [][][][]float64{}
	for _, polygon := range coords {
		newpolygon := [][][]float64{}
		for _, cont := range polygon {
			newline := [][]float64{}
			for _, pt := range cont {
				newline = append(newline, converter.Convert_XY(pt))
			}
			newpolygon = append(newpolygon, newline)
		}
		total = append(total, newpolygon)
	}
	return total
}

func Convert_Point_Merc(point []float64) []float64 {
	x := mercatorPole / 180.0 * point[0]

	y := math.Log(math.Tan((90.0+point[1])*math.Pi/360.0)) / math.Pi * mercatorPole
	y = math.Max(-mercatorPole, math.Min(y, mercatorPole))
	return []float64{x, y}
}

// converting points to long,lat
func Convert_Merc_Point(point []float64) []float64 {
	x := float64(point[0]) / (math.Pi / 180.0) / 6378137.0
	y := 180.0 / math.Pi * (2.0*math.Atan(math.Exp((float64(point[1])/6378137.0))) - math.Pi/2.0)
	return []float64{x, y}
}

func New_Point_Convert(tileid m.TileID) Point_Convert {
	bds := Create_Mercator_Bounds(tileid)
	return Point_Convert{Bds: bds, DeltaX: bds.E - bds.W, DeltaY: bds.N - bds.S}
}

// creates mercator bounds
func Create_Mercator_Bounds(tileid m.TileID) m.Extrema {
	bounds := m.Bounds(tileid)
	en := []float64{bounds.E, bounds.N} // east, north point
	ws := []float64{bounds.W, bounds.S} // west, south point

	// converting these
	en = Convert_Point_Merc(en)
	ws = Convert_Point_Merc(ws)

	// gettting north east west south
	east := en[0]
	north := en[1]
	west := ws[0]
	south := ws[1]
	bounds = m.Extrema{N: north, E: east, S: south, W: west}
	return bounds
}

// decodes a given delta
func DecodeDelta(nume uint32) int {
	num := int(nume)
	if num%2 == 1 {
		return (num + 1) / -2
	} else {
		return num / 2
	}
}

func Get_Command_Length(cmdLen uint32) (int32, int32) {
	cmd := cmdLen & 0x7
	length := cmdLen >> 3

	return int32(cmd), int32(length)
}

// decoding point
func Decode_Point(geom []uint32) [][]int {
	pos := 0
	firstpt, currentpt := []int{}, []int{}
	boolval := false
	newline := [][]int{}
	for pos < len(geom) {
		// getting geometry
		geomval := geom[pos]
		cmd, length := Get_Command_Length(geomval)

		// processing either multi geometry or single geometry
		if cmd == 1 && length == 1 && boolval == false {
			xdelta := DecodeDelta(geom[pos+1])
			ydelta := DecodeDelta(geom[pos+2])
			firstpt = []int{xdelta, ydelta}
			currentpt = firstpt
			boolval = true
			newline = append(newline, firstpt)

		} else if cmd == 1 {
			count := 1
			xdelta := DecodeDelta(geom[pos+1])
			ydelta := DecodeDelta(geom[pos+2])
			firstpt = []int{xdelta, ydelta}
			currentpt = firstpt
			newline = append(newline, currentpt)
			pos += 2
			for count < int(length) {
				xdelta := DecodeDelta(geom[pos+1])
				ydelta := DecodeDelta(geom[pos+2])
				currentpt = []int{currentpt[0] + xdelta, currentpt[1] + ydelta}
				newline = append(newline, currentpt)
				pos += 2
				count += 1
			}
		}
		pos += 1
	}
	return newline
}

// decodes a given simple geometry
// returns it as an arbitary set of given lines
func Decode_Line(geom []uint32) [][][]int {
	pos := 0
	currentpt := []int{0, 0}
	newline := [][]int{}
	lines := [][][]int{}
	for pos < len(geom) {
		geomval := geom[pos]

		cmd, length := Get_Command_Length(geomval)

		// conde for a move to cmd
		if cmd == 1 {
			xdelta := DecodeDelta(geom[pos+1])
			ydelta := DecodeDelta(geom[pos+2])
			currentpt = []int{currentpt[0] + xdelta, currentpt[1] + ydelta}
			pos += 2

			if pos == len(geom)-1 {
				lines = append(lines, [][]int{currentpt})
			}
		} else if cmd == 2 {
			newline = [][]int{currentpt}
			currentpos := pos + 1
			endpos := currentpos + int(length*2)
			for currentpos < endpos {
				xdelta := DecodeDelta(geom[currentpos])
				ydelta := DecodeDelta(geom[currentpos+1])
				currentpt = []int{currentpt[0] + xdelta, currentpt[1] + ydelta}
				newline = append(newline, currentpt)
				currentpos += 2
			}

			pos = currentpos - 1
			lines = append(lines, newline)

		}
		newline = [][]int{}
		//fmt.Println(cmd,length)
		pos += 1
	}
	return lines
}

// returns a bool on whether this is an exterior ring or not
func Exterior_Ring(coord [][]int) bool {
	count := 0
	firstpt := coord[0]
	weight := 0.0
	var oldpt []int
	for _, pt := range coord {
		if count == 0 {
			count = 1
		} else {
			weight += float64((pt[0] - oldpt[0]) * (pt[1] + oldpt[1]))
		}
		oldpt = pt
	}

	weight += float64((firstpt[0] - oldpt[0]) * (firstpt[1] + oldpt[1]))
	return weight > 0
}

// decodes a given simple geometry
// returns it as an arbitary set of given lines
func Decode_Polygon(geom []uint32) [][][][]int {
	pos := 0
	currentpt := []int{0, 0}
	newline := [][]int{}
	polygons := [][][][]int{}
	for pos < len(geom) {
		geomval := geom[pos]

		cmd, length := Get_Command_Length(geomval)

		// conde for a move to cmd
		if cmd == 1 {
			xdelta := DecodeDelta(geom[pos+1])
			ydelta := DecodeDelta(geom[pos+2])
			currentpt = []int{currentpt[0] + xdelta, currentpt[1] + ydelta}
			//fmt.Println(firstpt)
			pos += 2

		} else if cmd == 2 {
			newline = [][]int{currentpt}
			currentpos := pos + 1
			endpos := currentpos + int(length*2)
			for currentpos < endpos {
				xdelta := DecodeDelta(geom[currentpos])
				ydelta := DecodeDelta(geom[currentpos+1])
				currentpt = []int{currentpt[0] + xdelta, currentpt[1] + ydelta}
				newline = append(newline, currentpt)
				currentpos += 2
			}

			pos = currentpos - 1

		} else if cmd == 7 {
			//newline = append(newline,newline[0])
			if Exterior_Ring(newline) == false {
				polygons = append(polygons, [][][]int{newline})
				newline = [][]int{}
			} else {
				if len(polygons) == 0 {
					polygons = append(polygons, [][][]int{newline})

				} else {
					polygons[len(polygons)-1] = append(polygons[len(polygons)-1], newline)

				}
				newline = [][]int{}
			}

		}

		//fmt.Println(cmd,length)
		pos += 1
	}

	return polygons
}

// decodes a given simple geometry
// returns it as an arbitary set of given lines
func DecodeGeometry(geom []uint32) [][][]int {
	pos := 0
	firstpt, currentpt := []int{}, []int{}
	newline := [][]int{}
	lines := [][][]int{}
	for pos < len(geom) {
		geomval := geom[pos]

		cmd, length := Get_Command_Length(geomval)

		// conde for a move to cmd
		if cmd == 1 {
			xdelta := DecodeDelta(geom[pos+1])
			ydelta := DecodeDelta(geom[pos+2])
			firstpt = []int{xdelta, ydelta}
			currentpt = firstpt
			//fmt.Println(firstpt)
			pos += 2

			if pos == len(geom)-1 {
				lines = append(lines, [][]int{currentpt})
			}
		} else if cmd == 2 {
			newline = [][]int{firstpt}
			currentpos := pos + 1
			endpos := currentpos + int(length*2)
			for currentpos < endpos {
				xdelta := DecodeDelta(geom[currentpos])
				ydelta := DecodeDelta(geom[currentpos+1])
				currentpt = []int{currentpt[0] + xdelta, currentpt[1] + ydelta}
				newline = append(newline, currentpt)
				currentpos += 2
			}

			pos = currentpos - 1
			lines = append(lines, newline)

		} else if cmd == 7 {
			newline := lines[len(lines)-1]
			newline = append(newline, newline[0])
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
func Make_Key_Value_Map(values []*vector_tile.Tile_Value, keys []string) (map[int]interface{}, map[int]string) {
	// making value map
	valuemap := map[int]interface{}{}
	for i, value := range values {
		valuemap[i] = Get_Value(value)
	}

	// making keymap
	keymap := map[int]string{}
	for i, key := range keys {
		keymap[i] = key
	}

	return valuemap, keymap
}

// function to decode an entire set of geometries and return a list of geoms
func Decode_Geometry(feat_type vector_tile.Tile_GeomType, geom []uint32, converter Point_Convert) []*geojson.Geometry {
	geoms := []*geojson.Geometry{}
	if feat_type == vector_tile.Tile_POLYGON {
		coords := converter.Convert_Coords(Decode_Polygon(geom))
		for _, polygon := range coords {
			geoms = append(geoms, &geojson.Geometry{Type: "Polygon", Polygon: polygon})
		}
	} else if feat_type == vector_tile.Tile_LINESTRING {
		coords := converter.Convert_Coords([][][][]int{Decode_Line(geom)})
		for _, line := range coords[0] {
			geoms = append(geoms, &geojson.Geometry{Type: "LineString", LineString: line})
		}
	} else if feat_type == vector_tile.Tile_POINT {
		coords := converter.Convert_Coords([][][][]int{[][][]int{Decode_Point(geom)}})
		for _, point := range coords[0][0] {
			geoms = append(geoms, &geojson.Geometry{Type: "Point", Point: point})
		}
	}
	return geoms
}

func Convert_Offset(coords [][][][]int, k m.TileID) [][][][]int {
	///k.Y = (1 << uint64(k.Z)) - 1 - k.Y
	//offsety := k.Y * 4096
	//offsetx := k.X * 4096
	for polygoni := range coords {
		for conti := range coords[polygoni] {
			for pti := range coords[polygoni][conti] {
				pt := coords[polygoni][conti][pti]
				pt[0] = pt[0] >> 8
				pt[1] = pt[1] >> 8
				//pt[0] = int(math.Sqrt(float64(pt[0])))
				//pt[1] = int(math.Sqrt(float64(pt[1])))
				coords[polygoni][conti][pti] = pt
			}
		}

	}
	//fmt.Println(coords)
	return coords
}

// function to decode an entire set of geometries and return a list of geoms
func Decode_Geometry2(feat_type vector_tile.Tile_GeomType, geom []uint32, converter Point_Convert, k m.TileID) []*geojson.Geometry {
	geoms := []*geojson.Geometry{}
	if feat_type == vector_tile.Tile_POLYGON {
		coords := converter.Convert_Coords(Convert_Offset(Decode_Polygon(geom), k))
		for _, polygon := range coords {
			geoms = append(geoms, &geojson.Geometry{Type: "Polygon", Polygon: polygon})
		}
	} else if feat_type == vector_tile.Tile_LINESTRING {
		coords := converter.Convert_Coords(Convert_Offset([][][][]int{Decode_Line(geom)}, k))
		for _, line := range coords[0] {
			geoms = append(geoms, &geojson.Geometry{Type: "LineString", LineString: line})
		}
	} else if feat_type == vector_tile.Tile_POINT {
		coords := converter.Convert_Coords(Convert_Offset([][][][]int{[][][]int{Decode_Point(geom)}}, k))
		for _, point := range coords[0][0] {
			geoms = append(geoms, &geojson.Geometry{Type: "Point", Point: point})
		}
	}
	return geoms
}

// returns a list of features associated with each vector tile
func Convert_Vt_Bytes(bytevals []byte, tileid m.TileID) map[string][]*geojson.Feature {
	tile := &vector_tile.Tile{}
	if err := proto.Unmarshal(bytevals, tile); err != nil {
		fmt.Println(err)
	}

	converter := New_Point_Convert(tileid)
	newmap := map[string][]*geojson.Feature{}
	for _, layer := range tile.Layers {
		// getting value and key map
		valuemap, keymap := Make_Key_Value_Map(layer.Values, layer.Keys)
		//fmt.Println(valuemap, keymap)

		newfeats := []*geojson.Feature{}

		// making channel
		c := make(chan []*geojson.Feature)

		//converting coordinates
		for _, feat := range layer.Features {
			go func(feat *vector_tile.Tile_Feature, c chan []*geojson.Feature) {
				geoms := Decode_Geometry(*feat.Type, feat.Geometry, converter)
				/*
					var geom *geojson.Geometry
					coords := converter.Convert_Coords(DecodeGeometry(feat.Geometry))
					if *feat.Type == vector_tile.Tile_POLYGON {
						geom = &geojson.Geometry{Polygon:coords,Type:"Polygon"}
					} else if *feat.Type == vector_tile.Tile_LINESTRING {
						geom = &geojson.Geometry{LineString:coords[0],Type:"LineString"}
					} else if *feat.Type == vector_tile.Tile_POINT {
						geom = &geojson.Geometry{Point:coords[0][0],Type:"Point"}
					}
				*/
				properties := map[string]interface{}{}
				count := 0
				for count < len(feat.Tags) {
					keyid := int(feat.Tags[count])
					valueid := int(feat.Tags[count+1])
					key, value := keymap[keyid], valuemap[valueid]
					properties[key] = value
					count += 2
				}
				tempfeats := []*geojson.Feature{}
				for _, geom := range geoms {
					tempfeats = append(tempfeats, &geojson.Feature{ID: feat.Id, Geometry: geom, Properties: properties})
				}

				c <- tempfeats
			}(feat, c)
		}

		for range layer.Features {
			newfeats = append(newfeats, <-c...)
		}

		newmap[*layer.Name] = newfeats

	}
	return newmap
}

// returns a list of features associated with each vector tile
func Convert_Vt_Bytes_QA(bytevals []byte, tileid m.TileID) map[string][]*geojson.Feature {
	bytevals,err := GUnzipData(bytevals)
	if err != nil {
		fmt.Println(err)
	}

	tile := &vector_tile.Tile{}
	if err := proto.Unmarshal(bytevals, tile); err != nil {
		fmt.Println(err)
	}

	converter := New_Point_Convert(tileid)
	newmap := map[string][]*geojson.Feature{}
	for _, layer := range tile.Layers {
		// getting value and key map
		valuemap, keymap := Make_Key_Value_Map(layer.Values, layer.Keys)
		newfeats := []*geojson.Feature{}

		// making channel
		c := make(chan []*geojson.Feature)

		//converting coordinates
		for _, feat := range layer.Features {
			go func(feat *vector_tile.Tile_Feature, c chan []*geojson.Feature) {
				geoms := Decode_Geometry2(*feat.Type, feat.Geometry, converter, tileid)
				/*
					var geom *geojson.Geometry
					coords := converter.Convert_Coords(DecodeGeometry(feat.Geometry))
					if *feat.Type == vector_tile.Tile_POLYGON {
						geom = &geojson.Geometry{Polygon:coords,Type:"Polygon"}
					} else if *feat.Type == vector_tile.Tile_LINESTRING {
						geom = &geojson.Geometry{LineString:coords[0],Type:"LineString"}
					} else if *feat.Type == vector_tile.Tile_POINT {
						geom = &geojson.Geometry{Point:coords[0][0],Type:"Point"}
					}
				*/
				properties := map[string]interface{}{}
				count := 0
				for count < len(feat.Tags) {
					keyid := int(feat.Tags[count])
					valueid := int(feat.Tags[count+1])
					key, value := keymap[keyid], valuemap[valueid]
					properties[key] = value
					count += 2
				}
				tempfeats := []*geojson.Feature{}
				for _, geom := range geoms {
					tempfeats = append(tempfeats, &geojson.Feature{ID: feat.Id, Geometry: geom, Properties: properties})
				}

				c <- tempfeats
			}(feat, c)
		}

		for range layer.Features {
			newfeats = append(newfeats, <-c...)
		}

		newmap[*layer.Name] = newfeats

	}
	return newmap
}
