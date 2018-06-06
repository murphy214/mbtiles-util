package drawer

import (
	m "github.com/murphy214/mercantile"
	//pc "github.com/murphy214/polyclip"
	//"fmt"
	"github.com/paulmach/go.geojson"
	"math"
	"sort"
)

// BoundingBox implementation as per https://tools.ietf.org/html/rfc7946
// BoundingBox syntax: "bbox": [west, south, east, north]
// BoundingBox defaults "bbox": [-180.0, -90.0, 180.0, 90.0]
func BoundingBox_Points(pts [][]float64) []float64 {
	// setting opposite default values
	west, south, east, north := 180.0, 90.0, -180.0, -90.0

	for _, pt := range pts {
		x, y := pt[0], pt[1]
		// can only be one condition
		// using else if reduces one comparison
		if x < west {
			west = x
		} else if x > east {
			east = x
		}

		if y < south {
			south = y
		} else if y > north {
			north = y
		}
	}
	return []float64{west, south, east, north}
}

func Push_Two_BoundingBoxs(bb1 []float64, bb2 []float64) []float64 {
	// setting opposite default values
	west, south, east, north := 180.0, 90.0, -180.0, -90.0

	// setting bb1 and bb2
	west1, south1, east1, north1 := bb1[0], bb1[1], bb1[2], bb1[3]
	west2, south2, east2, north2 := bb2[0], bb2[1], bb2[2], bb2[3]

	// handling west values: min
	if west1 < west2 {
		west = west1
	} else {
		west = west2
	}

	// handling south values: min
	if south1 < south2 {
		south = south1
	} else {
		south = south2
	}

	// handling east values: max
	if east1 > east2 {
		east = east1
	} else {
		east = east2
	}

	// handling north values: max
	if north1 > north2 {
		north = north1
	} else {
		north = north2
	}

	return []float64{west, south, east, north}
}

// this functions takes an array of bounding box objects and
// pushses them all out
func Expand_BoundingBoxs(bboxs [][]float64) []float64 {
	bbox := bboxs[0]
	for _, temp_bbox := range bboxs[1:] {
		bbox = Push_Two_BoundingBoxs(bbox, temp_bbox)
	}
	return bbox
}

// boudning box on a normal point geometry
// relatively useless
func BoundingBox_PointGeometry(pt []float64) []float64 {
	return []float64{pt[0], pt[1], pt[0], pt[1]}
}

// Returns BoundingBox for a MultiPoint
func BoundingBox_MultiPointGeometry(pts [][]float64) []float64 {
	return BoundingBox_Points(pts)
}

// Returns BoundingBox for a LineString
func BoundingBox_LineStringGeometry(line [][]float64) []float64 {
	return BoundingBox_Points(line)
}

// Returns BoundingBox for a MultiLineString
func BoundingBox_MultiLineStringGeometry(multiline [][][]float64) []float64 {
	bboxs := [][]float64{}
	for _, line := range multiline {
		bboxs = append(bboxs, BoundingBox_Points(line))
	}
	return Expand_BoundingBoxs(bboxs)
}

// Returns BoundingBox for a Polygon
func BoundingBox_PolygonGeometry(polygon [][][]float64) []float64 {
	bboxs := [][]float64{}
	for _, cont := range polygon {
		bboxs = append(bboxs, BoundingBox_Points(cont))
	}
	return Expand_BoundingBoxs(bboxs)
}

// Returns BoundingBox for a Polygon
func BoundingBox_MultiPolygonGeometry(multipolygon [][][][]float64) []float64 {
	bboxs := [][]float64{}
	for _, polygon := range multipolygon {
		for _, cont := range polygon {
			bboxs = append(bboxs, BoundingBox_Points(cont))
		}
	}
	return Expand_BoundingBoxs(bboxs)
}

// Returns a BoundingBox for a geometry collection
func BoundingBox_GeometryCollection(gs []*geojson.Geometry) []float64 {
	bboxs := [][]float64{}
	for _, g := range gs {
		bboxs = append(bboxs, g.Get_BoundingBox())
	}
	return Expand_BoundingBoxs(bboxs)
}

// retrieves a boundingbox for a given geometry
func Get_BoundingBox(g *geojson.Geometry) []float64 {
	switch g.Type {
	case "Point":
		return BoundingBox_PointGeometry(g.Point)
	case "MultiPoint":
		return BoundingBox_MultiPointGeometry(g.MultiPoint)
	case "LineString":
		return BoundingBox_LineStringGeometry(g.LineString)
	case "MultiLineString":
		return BoundingBox_MultiLineStringGeometry(g.MultiLineString)
	case "Polygon":
		return BoundingBox_PolygonGeometry(g.Polygon)
	case "MultiPolygon":
		return BoundingBox_MultiPolygonGeometry(g.MultiPolygon)

	}
	return []float64{}
}

//
func GetBoundingingBox(g *geojson.Feature) m.Extrema {
	bb := Get_BoundingBox(g.Geometry)
	return m.Extrema{W: bb[0], S: bb[1], E: bb[2], N: bb[3]}
}

func GetSmallestTile(bds m.Extrema) m.TileID {
	corners := [][]float64{{bds.E, bds.N}, {bds.E, bds.S}, {bds.W, bds.S}, {bds.W, bds.N}}
	var tileid m.TileID
	boolval := false
	for i := 20; i >= 0; i-- {
		mymap := map[m.TileID]string{}
		for _, corner := range corners {
			mymap[m.Tile(corner[0], corner[1], i)] = ""
		}
		if len(mymap) == 1 && !boolval {
			boolval = true
			for k := range mymap {
				tileid = k
			}
		}
	}
	return tileid
}

// this data structure is specifically made to take a feature (line oro polygon to gridded data)
type Grid struct {
	Bds        m.Extrema
	DeltaX     float64
	DeltaY     float64
	Resolution int
	Size       int
	SizeX      int
	SizeY      int
	Screen     *Screen
}

func (grid *Grid) Border() [][2]int {
	current := 0
	newlist := []int{}
	for current < grid.Size {
		newlist = append(newlist, current)
		current += 1
	}

	// x = 0
	pixels := [][2]int{}
	for _, i := range newlist {
		pixels = append(pixels, [2]int{0, i})
	}

	for _, i := range newlist {
		pixels = append(pixels, [2]int{grid.Size - 1, i})
	}

	for _, i := range newlist {
		pixels = append(pixels, [2]int{i, 0})
	}
	for _, i := range newlist {
		pixels = append(pixels, [2]int{i, grid.Size - 1})
	}
	return pixels
}

// creates a single grid object
func New_Grid2(feature *geojson.Feature, resolution int) Grid {
	bb := GetBoundingingBox(feature)
	tileid := GetSmallestTile(bb)
	deltax := (bb.E - bb.W)
	deltay := (bb.N - bb.S)
	size := resolution
	bds := m.Bounds(tileid)
	deltax2 := (bds.E - bds.W)
	deltay2 := (bds.N - bds.S)
	ratiox := deltax / deltax2
	ratioy := deltay / deltay2
	var sizex, sizey int
	if ratiox > ratioy {
		sizex = size
		sizey = int(ratioy / ratiox * float64(size))
	} else {
		sizey = size
		sizex = int(ratiox / ratioy * float64(size))
	}
	//fmt.Println(sizex, sizey)
	deltax = (bb.E - bb.W) / float64(sizex)
	deltay = (bb.N - bb.S) / float64(sizey)
	return Grid{Bds: bb, Size: size, DeltaY: deltay, DeltaX: deltax, SizeX: sizex, SizeY: sizey}
}

// creates a single grid object
func NewGrid(tileid m.TileID, resolution int) Grid {
	bds := m.Bounds(tileid)
	size := resolution
	deltax := (bds.E - bds.W) / float64(size)
	deltay := (bds.N - bds.S) / float64(size)
	screen := NewScreen(resolution, resolution)
	return Grid{Bds: bds, Size: size, SizeX: size, SizeY: size, DeltaY: deltay, DeltaX: deltax, Screen: screen}
}

// grid to bds
func (grid *Grid) Pt_Extrema(pt [2]int) m.Extrema {
	long := float64(pt[0])*grid.DeltaX + grid.Bds.W
	lat := grid.Bds.N - float64(pt[1])*grid.DeltaY
	return m.Extrema{W: long - grid.DeltaX, E: long + grid.DeltaX, S: lat - grid.DeltaY, N: lat + grid.DeltaY}
}

// rounds a number to an adequate value
func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// grids sa. single point
func (grid *Grid) Grid_Pt(pt []float64) [2]int {
	//fmt.Println((pt[1] - grid.Bds.S))
	return [2]int{int(
		Round(math.Floor((pt[0]-grid.Bds.W)/grid.DeltaX), .5, 0),
	),
		int(
			Round(float64(grid.SizeY)-math.Ceil((pt[1]-grid.Bds.S)/grid.DeltaY), .5, 0),
		)}
}

func (grid *Grid) Pixel_Long(x int) float64 {
	return float64(x)*grid.DeltaX + grid.Bds.W
}

func (grid *Grid) Pixel_Lat(y int) float64 {
	return grid.Bds.N - float64(y)*grid.DeltaY
}

func Fill_Vals(int1, int2 int) []int {
	vals := []int{int1}
	current := int1
	for current < int2 {
		current += 1
		vals = append(vals, current)
	}
	if vals[len(vals)-1] == int2 {
		vals = append(vals, int2)
	}
	return vals
}

func Unique(vals []float64) []float64 {
	floatmap := map[float64]string{}
	newlist := []float64{}
	for _, val := range vals {
		_, boolval := floatmap[val]
		if boolval == false {
			newlist = append(newlist, val)
			floatmap[val] = ""
		}
	}
	return newlist
}

func Reverse(vals []int) []int {
	count := len(vals) - 1
	newvals := make([]int, len(vals))
	count2 := 0
	for count != -1 {
		newvals[count2] = vals[count]
		count -= 1
		count2 += 1

	}
	return newvals
}

//
func (grid *Grid) Make_Xs_Ys(bds_pixel m.Extrema) ([]float64, []float64) {
	//deltax := bds_pixel.N - bds_pixel.S
	//floatmap := map[float64]string{}
	xs := []float64{}
	for _, x := range Fill_Vals(int(bds_pixel.W), int(bds_pixel.E)) {
		long := grid.Pixel_Long(x)
		if x == int(bds_pixel.W) {
			xs = append(xs, long+grid.DeltaX)
			xs = append(xs, long)

		} else if x == int(bds_pixel.E) {
			xs = append(xs, long-grid.DeltaX)
			xs = append(xs, long)

		} else {
			xs = append(xs, []float64{long - grid.DeltaX + .0000001, long + grid.DeltaX - +.0000001}...)
		}
	}
	xs = Unique(xs)

	ys := []float64{}
	for _, y := range Reverse(Fill_Vals(int(bds_pixel.S), int(bds_pixel.N))) {
		lat := grid.Pixel_Lat(y)
		if y == int(bds_pixel.S) {
			ys = append(ys, lat-grid.DeltaY-.0000001)
			ys = append(ys, lat)

		} else if y == int(bds_pixel.N) {
			ys = append(ys, lat+grid.DeltaY+.0000001)
			ys = append(ys, lat)

		} else {
			ys = append(ys, []float64{lat - grid.DeltaY + .0000001, lat + grid.DeltaY - .0000001}...)
		}
	}

	ys = Unique(ys)
	return xs, ys
}

// regular interpolation
func Interpolate(pt1, pt2 []float64, x float64) []float64 {
	slope := (pt2[1] - pt1[1]) / (pt2[0] - pt1[0])
	return []float64{x, (x-pt1[0])*slope + pt1[1]}
}

// iterpolate the pixels between two points
func (grid *Grid) Interpolate(pt1 []float64, pt2 []float64) [][2]int {
	pixel1 := grid.Grid_Pt(pt1)
	pixel2 := grid.Grid_Pt(pt2)
	if pixel1[0] == pixel2[0] && pixel1[1] != pixel2[1] {
		vals := []int{pixel1[1], pixel2[1]}
		sort.Ints(vals)
		ys := Fill_Vals(vals[0], vals[1])
		pixels := make([][2]int, len(ys))
		for pos, y := range ys {
			pixels[pos] = [2]int{pixel1[0], y}
		}
		return pixels
	}
	if pixel1[0] != pixel2[0] && pixel1[1] == pixel2[1] {
		vals := []int{pixel1[0], pixel2[0]}
		sort.Ints(vals)
		xs := Fill_Vals(vals[0], vals[1])
		pixels := make([][2]int, len(xs))
		for pos, x := range xs {
			pixels[pos] = [2]int{x, pixel1[1]}
		}
		return pixels
	}
	if pixel1[0] == pixel2[0] && pixel1[1] == pixel2[1] {
		return [][2]int{pixel1}
	}

	var n, s, e, w int
	if pixel1[0] >= pixel2[0] {
		e, w = pixel1[0], pixel2[0]
	} else {
		w, e = pixel1[0], pixel2[0]
	}

	if pixel1[1] >= pixel2[1] {
		n, s = pixel1[1], pixel2[1]
	} else {
		s, n = pixel1[1], pixel2[1]
	}

	// getting the bds of the two pixels
	bds_pixel := m.Extrema{N: float64(n), S: float64(s), E: float64(e), W: float64(w)}
	//fmt.Println(bds_pixel)
	//fmt.Println(bds_pixel)
	xs, ys := grid.Make_Xs_Ys(bds_pixel)
	//fmt.Println(xs,ys,"ds")
	//fmt.Println(ys)

	mymap := map[[2]int]string{}
	for _, x := range xs {
		pt := Interpolate(pt1, pt2, x)
		pixel := grid.Grid_Pt(pt)
		if (bds_pixel.E >= float64(pixel[0])) && (bds_pixel.W <= float64(pixel[0])) && (bds_pixel.N >= float64(pixel[1])) && (bds_pixel.S <= float64(pixel[1])) {
			mymap[pixel] = ""

		}
	}

	pt1b := []float64{pt1[1], pt1[0]}
	pt2b := []float64{pt2[1], pt2[0]}
	//fmt.Println(ys)
	for _, y := range ys {

		pt := Interpolate(pt1b, pt2b, y)
		pt = []float64{pt[1], pt[0]}
		pixel := grid.Grid_Pt(pt)

		//fmt.Println(pixel)
		if (bds_pixel.E >= float64(pixel[0])) && (bds_pixel.W <= float64(pixel[0])) && (bds_pixel.N >= float64(pixel[1])) && (bds_pixel.S <= float64(pixel[1])) {
			mymap[pixel] = ""

		}

	}
	newlist := [][2]int{}
	//spec_size := grid.Size
	mymap[pixel1] = ""
	mymap[pixel2] = ""
	for k := range mymap {

		//if (k[0] >= 0) && (k[0] < spec_size) && (k[1] >= 0) && (k[1] < spec_size)  {
		newlist = append(newlist, [2]int{k[0], k[1]})
		//}
	}

	return newlist
}

// gets all the pixels along a line within an image
func (grid *Grid) Interpolate_Line(line [][]float64) [][2]int {
	count := 0
	pixels := [][2]int{}
	var oldpt []float64
	for _, pt := range line {
		pixel1 := grid.Grid_Pt(pt)
		//pixel1[1] = int(math.Abs(float64(pixel1[1])))

		pixels = append(pixels, pixel1)
		if count == 0 {
			count = 1
		} else {
			pixels = append(pixels, grid.Interpolate(oldpt, pt)...)
		}
		oldpt = pt
	}
	//Draw(pixels)
	return pixels
}

func Make_Map(newlist [][2]int) map[[2]int]string {
	mymap := map[[2]int]string{}
	for _, pt := range newlist {
		mymap[pt] = ""
	}
	return mymap
}

type Poly [][][]float64

func (cont Poly) Pip(p []float64) bool {
	// Cast ray from p.x towards the right
	intersections := 0
	for _, c := range cont {
		for i := range c {
			curr := c[i]
			ii := i + 1
			if ii == len(c) {
				ii = 0
			}
			next := c[ii]

			// Is the point out of the edge's bounding box?
			// bottom vertex is inclusive (belongs to edge), top vertex is
			// exclusive (not part of edge) -- i.e. p lies "slightly above
			// the ray"
			bottom, top := curr, next
			if bottom[1] >= top[1] {
				bottom, top = top, bottom
			}
			if p[1] <= bottom[1] || p[1] >= top[1] {
				continue
			}
			// Edge is from curr to next.

			if p[0] >= math.Max(curr[0], next[0]) ||
				next[1] == curr[1] {
				continue
			}

			// Find where the line intersects...
			xint := (p[1]-curr[1])*(next[0]-curr[0])/(next[1]-curr[1]) + curr[0]
			if curr[0] != next[0] && p[0] > xint {
				continue
			}

			intersections++
		}
	}

	return intersections%2 != 0
}

func Poly2(poly [][][]float64) interface{} {
	return poly
}

func Fix_Polygon(coords [][][]float64) [][][]float64 {
	for i, cont := range coords {
		if (cont[0][0] != cont[len(cont)-1][0]) && (cont[0][1] != cont[len(cont)-1][1]) {
			cont = append(cont, cont[0])
			coords[i] = cont
		}
	}
	return coords
}

func (grid *Grid) Interpolate_Polygon(polygon [][][]float64) ([][2]int, [][2]int) {
	raycast := map[int][]int{}
	holeraycast := map[int][]int{}
	polygon = Fix_Polygon(polygon)
	line_pixels := [][2]int{}
	for i, cont := range polygon {
		//if (cont[0][0] != cont[len(cont) - 1][0]) && (cont[0][1] != cont[len(cont) - 1][1]) {
		//	cont = append(cont,cont[0])
		//}
		cont = append(cont, cont[0])

		pixels := grid.Interpolate_Line(cont)
		line_pixels = append(line_pixels, pixels...)
		if i == 0 {
			for _, pt := range pixels {
				raycast[pt[0]] = append(raycast[pt[0]], pt[1])
			}
		} else {
			for _, pt := range pixels {
				holeraycast[pt[0]] = append(holeraycast[pt[0]], pt[1])
			}
		}
	}
	clip_polygon := Poly(polygon)
	//clip_polygon := clip_polygon2.(Poly)

	newlist := [][2]int{}
	for k, v := range raycast {
		count := 0
		var oldi int
		sort.Ints(v)
		//fmt.Println(v)
		for _, i := range v {

			if count == 0 {
				newlist = append(newlist, [2]int{k, i})
				count = 1

			} else {
				if i-oldi > 1 {
					// /avgi := (oldi + i) / 2
					//fmt.Println(avgi,i,oldi,"average")
					current := oldi
					count2 := 0
					for current < i {
						if count2 == 0 {
							newlist = append(newlist, [2]int{k, current})
						} else {

							pt := []float64{grid.Pixel_Long(k), grid.Pixel_Lat(current)}

							if clip_polygon.Pip(pt) {
								newlist = append(newlist, [2]int{k, current})
							}
						}
						current += 1
						count2 += 1

					}
				} else {
					newlist = append(newlist, [2]int{k, oldi})
				}
			}

			oldi = i
			//fmt.Println(k,v[i])
		}
	}

	map1 := Make_Map(newlist)

	pixels := [][2]int{}
	spec_size := grid.Size
	for k := range map1 {
		if (k[0] >= 0) && (k[0] < spec_size) && (k[1] >= 0) && (k[1] < spec_size) {
			pixels = append(pixels, k)
		}
	}

	return pixels, line_pixels
}
