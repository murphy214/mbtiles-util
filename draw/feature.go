package drawer

import (
	//"fmt"
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	"io/ioutil"
	"math"
	"math/rand"
)

// generates a random color integer
func RandColor() int {
	return brights[rand.Intn(len(brights)-1)]
}

// draws the pixels associated with one pixel
func (grid *Grid) PaintFeature(feature *geojson.Feature) ([][2]int, int) {
	switch feature.Geometry.Type {
	case "LineString":
		return grid.Interpolate_Line(feature.Geometry.LineString), RandColor()
	case "MultiLineString":
		pts := [][2]int{}
		for _, line := range feature.Geometry.MultiLineString {
			pts = append(pts, grid.Interpolate_Line(line)...)
		}
		return pts, RandColor()
	case "Polygon":
		pixels, _ := grid.Interpolate_Polygon(feature.Geometry.Polygon)
		return pixels, RandColor()
	case "MultiPolygon":
		pts := [][2]int{}

		for _, polygon := range feature.Geometry.MultiPolygon {
			tmppixels, _ := grid.Interpolate_Polygon(polygon)
			pts = append(pts, tmppixels...)
		}
		return pts, RandColor()
	}

	return [][2]int{}, 0
}

// draws the pixels associated with one pixel
func (grid *Grid) DrawFeature(feature *geojson.Feature) ([][2]int, [][2]int) {
	switch feature.Geometry.Type {
	case "LineString":
		return grid.Interpolate_Line(feature.Geometry.LineString), [][2]int{}
	case "MultiLineString":
		pts := [][2]int{}
		for _, line := range feature.Geometry.MultiLineString {
			pts = append(pts, grid.Interpolate_Line(line)...)
		}
		return pts, [][2]int{}
	case "Polygon":
		pixels, bdr_pixels := grid.Interpolate_Polygon(feature.Geometry.Polygon)

		return pixels, bdr_pixels
	case "MultiPolygon":
		pts := [][2]int{}
		bdr_pts := [][2]int{}
		for _, polygon := range feature.Geometry.MultiPolygon {
			tmppixels, tmppixels2 := grid.Interpolate_Polygon(polygon)
			pts = append(pts, tmppixels...)
			bdr_pts = append(bdr_pts, tmppixels2...)
		}
		return pts, bdr_pts
	}

	return [][2]int{}, [][2]int{}
}

// paints a given set of features to a screen
func (grid *Grid) PaintFeatures(features []*geojson.Feature) {
	for _, feature := range features {
		pts, colorint := grid.PaintFeature(feature)

		// setting all the points to that color integer
		for _, pt := range pts {
			//fmt.Println(pt)
			x, y := pt[0], pt[1]
			y = int(math.Abs(float64(y)))
			pt = [2]int{x, y}
			if x < grid.Screen.X && x > 0 && y < grid.Screen.Y && y > 0 {
				grid.Screen.Set(pt, colorint)
			}
		}
	}
}

func CreateFeature(feature *geojson.Feature, tileid m.TileID, dim int) {
	grid := NewGrid(tileid, 4095)
	pts, color := grid.PaintFeature(feature)
	minx, maxx, miny, maxy := 4095, 0, 4095, 0
	for i, pt := range pts {
		pt[1] = int(math.Abs(float64(pt[1])))
		//fmt.Println(pt)
		if pt[0] > maxx {
			maxx = pt[0]
		}

		if pt[0] < minx {
			minx = pt[0]
		}
		if pt[1] > maxy {
			maxy = pt[1]
		}

		if pt[1] < miny {
			miny = pt[1]
		}

		pts[i] = pt
	}

	deltax := maxx - minx
	deltay := maxy - miny

	var size int
	if deltax > deltay {
		size = deltax
	} else {
		size = deltay
	}
	scale := float64(dim) / float64(size)

	for pos, i := range pts {
		i = [2]int{int(Round(float64((i[0]-minx))*scale, .5, 0)), int(Round(float64((i[1]-miny))*scale, .5, 0))}
		pts[pos] = i
	}
	screen := NewScreen(dim, dim)
	for _, pt := range pts {
		//fmt.Println(pt)
		x, y := pt[0], pt[1]

		if x < screen.X && x > 0 && y < screen.Y && y > 0 {
			//fmt.Println(y+1, grid.Screen.Y)

			screen.Set(pt, color)
		}
	}
	screen.Draw()
}

func AbsPixels(pixels [][2]int, res int) [][2]int {
	newpixels := [][2]int{}
	for _, pixel := range pixels {
		pixel[1] = pixel[1] * -1
		if pixel[0] >= 0 && pixel[0] <= res && pixel[1] >= 0 && pixel[1] <= res {
			newpixels = append(newpixels, pixel)
		}
	}
	return newpixels
}

func WriteFeaturesPNG(features []*geojson.Feature, tileid m.TileID, res int, filename string) {
	img := NewImg(res, res)

	grid := NewGrid(tileid, res)

	for _, feature := range features {
		pixels, bdr := grid.DrawFeature(feature)
		//pixels, bdr = AbsPixels(pixels, res), AbsPixels(bdr, res)
		img.Pixels = append(img.Pixels, pixels)
		img.Line_Pixels = append(img.Line_Pixels, bdr...)
	}
	//fmt.Println(img)

	bytevals := img.Write()
	ioutil.WriteFile(filename, bytevals, 0677)
}

func WriteFeaturesMaskPNG(features []*geojson.Feature, tileid m.TileID, res int, filename string) {
	img := NewImg(res, res)

	grid := NewGrid(tileid, res)

	for _, feature := range features {
		pixels, bdr := grid.DrawFeature(feature)
		//pixels, bdr = AbsPixels(pixels, res), AbsPixels(bdr, res)
		img.Pixels = append(img.Pixels, pixels)
		img.Line_Pixels = append(img.Line_Pixels, bdr...)
	}
	//fmt.Println(img)

	bytevals := img.WriteMask()
	ioutil.WriteFile(filename, bytevals, 0677)
}
