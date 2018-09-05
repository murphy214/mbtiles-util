package drawer

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"strings"
)

var default_string = "\033[38;5;%dm\033[48;5;%dm▀\033[0m"
var brights = []int{9, 10, 11, 12, 13, 14, 47, 88, 90, 154}

type Screen struct {
	X      int
	Y      int
	Screen [][]int
}

func NewScreen(x, y int) *Screen {
	// /row := make([]int, x)
	screen := make([][]int, y)
	for i := range screen {
		screen[i] = make([]int, x)
	}
	return &Screen{Screen: screen, X: x, Y: y}
}

func (screen Screen) Draw() {
	totalrows := make([]string, screen.Y/2)
	for i := 0; i < screen.Y; i += 2 {
		newlist := make([]string, screen.X)
		for x := 0; x < screen.X; x++ {
			newlist[x] = fmt.Sprintf(default_string, screen.Screen[i][x], screen.Screen[i+1][x])
		}
		rowstring := strings.Join(newlist, "")
		totalrows[i/2] = rowstring
	}
	fmt.Println(strings.Join(totalrows, "\n"))
}
func (screen Screen) DrawReturn() string {
	totalrows := make([]string, screen.Y/2)
	for i := 0; i < screen.Y; i += 2 {
		newlist := make([]string, screen.X)
		for x := 0; x < screen.X; x++ {
			newlist[x] = fmt.Sprintf(default_string, screen.Screen[i][x], screen.Screen[i+1][x])
		}
		rowstring := strings.Join(newlist, "")
		totalrows[i/2] = rowstring
	}
	return strings.Join(totalrows, "\n")
}
func (screen *Screen) Set(pt [2]int, color int) {
	x, y := pt[0], pt[1]
	screen.Screen[y][x] = color
}

func RandomPt(x, y int) [2]int {
	return [2]int{rand.Intn(x), rand.Intn(y)}
}

func RandomPts(x, y, number int) [][2]int {
	pts := make([][2]int, number)
	for i := 0; i < number; i++ {

		pts[i] = RandomPt(x, y)
	}
	return pts

}

var bright_colors_uint8 = [][]uint8{{252, 73, 163}, {204, 102, 255}, {102, 204, 255}, {102, 255, 204}, {0, 255, 0}, {255, 204, 102}, {255, 102, 102}, {255, 0, 0}, {255, 128, 0}, {255, 255, 102}, {0, 255, 255}}

type Image_Mask struct {
	Pixels      [][][2]int
	Line_Pixels [][2]int
	X, Y        int
}

func NewImg(x, y int) *Image_Mask {
	return &Image_Mask{X: x, Y: y}
}

func (img *Image_Mask) Write() []byte {
	im := image.NewRGBA(image.Rectangle{Max: image.Point{X: img.X, Y: img.X}})

	for _, pixels := range img.Pixels {
		bright_colors := bright_colors_uint8[rand.Intn(len(bright_colors_uint8)-1)]
		c := color.RGBA{bright_colors[0], bright_colors[1], bright_colors[2], 255}
		for _, k := range pixels {
			im.SetRGBA(k[0], k[1], c)
		}
	}

	c := color.RGBA{0, 0, 0, 255}
	for _, k := range img.Line_Pixels {
		im.SetRGBA(k[0], k[1], c)
	}

	a := bytes.NewBuffer([]byte{})
	if err := png.Encode(a, im); err != nil {
		log.Fatal(err)
	}
	return a.Bytes()
}

func (img *Image_Mask) WriteMask() []byte {
	im := image.NewGray(image.Rectangle{Max: image.Point{X: img.X, Y: img.X}})

	for _, pixels := range img.Pixels {
		for _, k := range pixels {

			// adding two iterations of dialtion
			/*

				TO DO:
				This is really a hack and dialation should be handled in a more cohesive manner.

				CURRENTLY:

				X  X  X  X  X
				X  X  X  X  X
				X  X  O  X  X
				X  X  X  X  X
				X  X  X  X  X

				WHERE O is the subject point
			*/
			gridpt1 := [2]int{k[0] - 3, k[1] - 3}
			gridpt2 := [2]int{k[0] + 3, k[1] + 3}
			for currentx := gridpt1[0]; currentx < gridpt2[0]; currentx++ {
				for currenty := gridpt1[1]; currenty < gridpt2[1]; currenty++ {
					im.SetGray(currentx, currenty, color.Gray{uint8(255)})
				}
			}
			//im.SetGray(k[0], k[1], color.Gray{uint8(255)})
		}
	}
	/*
		c := color.Gray{uint8(1)}
		for _, k := range img.Line_Pixels {
			im.SetGray(k[0], k[1], c)
		}
	*/
	a := bytes.NewBuffer([]byte{})
	if err := png.Encode(a, im); err != nil {
		log.Fatal(err)
	}
	return a.Bytes()
}
