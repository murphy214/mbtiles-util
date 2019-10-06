package mbutil

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	//"os"
	"bytes"
	"compress/gzip"
	//"github.com/paulmach/go.geojson"
	"io"
	"sync"
)

// unzips a write
func GZipWrite(bs []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(bs)
	w.Flush()
	w.Close() // You must close this first to flush the bytes to the buffer.
	return b.Bytes()
}


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

// given a filename gets a filename structure to read
func ReadMbtiles(filename string) (Mbtiles, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return Mbtiles{}, err
	}
	// starting the transaction for adding tiles
	tx, err := db.Begin()
	if err != nil {
		return Mbtiles{}, err
	}

	var mutex sync.Mutex
	mb := Mbtiles{Tx: tx,
		Mutex:     mutex,
		NewBool:   false,
		Old_Total: -1,
		FileName:  filename}
	minzoom, maxzoom := mb.GetMinMaxZoom()
	mb.MinZoom = minzoom
	mb.MaxZoom = maxzoom
	return mb, nil

}

// queries a given tileid
func (mbtiles *Mbtiles) Query(k m.TileID) ([]byte, error) {
	k.Y = (1 << uint64(k.Z)) - 1 - k.Y
	var data []byte
	err := mbtiles.Tx.QueryRow("select tile_data from tiles where zoom_level = ? and tile_column = ? and tile_row = ?", k.Z, k.X, k.Y).Scan(&data)
	if err != nil {
		return []byte{}, err
	}
	
	if len(data) >= 2 {
		if (data[0] == 0x1f) && (data[1] == 0x8b) {
			data, err = GUnzipData(data)
			if err != nil {
				return []byte{}, err

			}
		}
	}
	return data, nil
}


// queries a given tileid returns an image
func (mbtiles *Mbtiles) QueryImage(k m.TileID) ([]byte, error) {
	k.Y = (1 << uint64(k.Z)) - 1 - k.Y
	var data []byte
	tileid := fmt.Sprintf("%d/%d/%d",k.Z,k.X,k.Y)
	err := mbtiles.Tx.QueryRow("select tile_data from images where tileid=?", tileid).Scan(&data)
	if err != nil {
		return []byte{}, err
	}
	
	if len(data) >= 2 {
		if (data[0] == 0x1f) && (data[1] == 0x8b) {
			data, err = GUnzipData(data)
			if err != nil {
				return []byte{}, err

			}
		}
	}
	return data, nil
}

// queries a given tileid
func (mbtiles *Mbtiles) QueryRaw(k m.TileID) ([]byte, error) {
	k.Y = (1 << uint64(k.Z)) - 1 - k.Y
	var data []byte
	err := mbtiles.Tx.QueryRow("select tile_data from tiles where zoom_level = ? and tile_column = ? and tile_row = ?", k.Z, k.X, k.Y).Scan(&data)
	if err != nil {
		return []byte{}, err
	}
	
	return data, nil
}

// function for getting the next value
func (mbtiles *Mbtiles) Next() bool {
	return (mbtiles.Old_Total == mbtiles.Total) == false
}

//
func (mbtiles *Mbtiles) GetMax() int {
	var myint int
	_ = mbtiles.Tx.QueryRow("SELECT COUNT(*) from tiles;").Scan(&myint)
	return myint
}

// map query
func (mbtiles *Mbtiles) MapTilesGet(limit int) map[m.TileID][]byte {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles limit ?,?", mbtiles.Total, limit)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := map[m.TileID][]byte{}
	step := 0
	for rows.Next() {
		var x, y, z int
		var data []byte

		rows.Scan(&x, &y, &z, &data)
		if (data[0] == 0x1f) && (data[1] == 0x8b) {
			data, err = GUnzipData(data)
			if err != nil {
				fmt.Println(err)
			}
		}
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap[tileid] = data
		step += 1
	}
	mbtiles.Old_Total = mbtiles.Total
	mbtiles.Total += step

	return mymap
}

// map query
func (mbtiles *Mbtiles) ChunkTiles(limit int) map[m.TileID][]byte {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles limit ?,?", mbtiles.Total, limit)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := map[m.TileID][]byte{}
	step := 0
	for rows.Next() {
		var x, y, z int
		var data []byte

		rows.Scan(&x, &y, &z, &data)
		if (data[0] == 0x1f) && (data[1] == 0x8b) {
			data, err = GUnzipData(data)
			if err != nil {
				fmt.Println(err)
			}
		}
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap[tileid] = data
		step += 1
	}
	mbtiles.Old_Total = mbtiles.Total
	mbtiles.Total += step

	return mymap
}

// map query
func (mbtiles *Mbtiles) ChunkTilesMeta(limit int) map[m.TileID]int {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,length(tile_data) FROM tiles limit ?,?", mbtiles.Total, limit)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := map[m.TileID]int{}
	step := 0
	for rows.Next() {
		var x, y, z int
		var data int

		rows.Scan(&x, &y, &z, &data)

		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap[tileid] = data
		step += 1
	}
	mbtiles.Old_Total = mbtiles.Total
	mbtiles.Total += step

	return mymap
}

// map query
func (mbtiles *Mbtiles) ChunkTilesZoom(limit int, zoom int) map[m.TileID][]byte {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles where zoom_level = ? limit ?,?", zoom, mbtiles.Total, limit)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := map[m.TileID][]byte{}
	step := 0
	for rows.Next() {
		var x, y, z int
		var data []byte

		rows.Scan(&x, &y, &z, &data)
		if (data[0] == 0x1f) && (data[1] == 0x8b) {
			data, err = GUnzipData(data)
			if err != nil {
				fmt.Println(err)
			}
		}
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap[tileid] = data
		step += 1
	}
	mbtiles.Old_Total = mbtiles.Total
	mbtiles.Total += step

	return mymap
}

// map query
func (mbtiles *Mbtiles) GetOneTile() (m.TileID, []byte) {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles limit ?,?", mbtiles.Total, mbtiles.Total+1)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := []m.TileID{}
	step := 0
	for rows.Next() {
		var x, y, z int
		var data []byte

		rows.Scan(&x, &y, &z, &data)
		if (data[0] == 0x1f) && (data[1] == 0x8b) {
			data, err = GUnzipData(data)
			if err != nil {
				fmt.Println(err)
			}
		}
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap = append(mymap, tileid)
		step += 1
	}

	tileid := mymap[0]
	bytevals, _ := mbtiles.Query(tileid)
	mbtiles.Total += 1

	return tileid, bytevals
}

func (mbtiles *Mbtiles) SingleTile() m.TileID {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles limit ?,?", 1, mbtiles.Total+1)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := []m.TileID{}
	step := 0
	for rows.Next() {
		var x, y, z int
		var data []byte

		rows.Scan(&x, &y, &z, &data)
		if (data[0] == 0x1f) && (data[1] == 0x8b) {
			data, err = GUnzipData(data)
			if err != nil {
				fmt.Println(err)
			}
		}
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap = append(mymap, tileid)
		step += 1
	}

	tileid := mymap[0]
	mbtiles.Total += 1

	return tileid
}

// map query
func (mbtiles *Mbtiles) GetAllTiles() []m.TileID {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level FROM tiles;")
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := []m.TileID{}
	step := 0
	for rows.Next() {
		var x, y, z int

		rows.Scan(&x, &y, &z)
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap = append(mymap, tileid)
		step += 1
	}
	return mymap
}

// map query
func (mbtiles *Mbtiles) GetAllTilesSorted() []m.TileID {
	rows, err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level FROM tiles ORDER BY length(tile_data) DESC;")
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid
	mymap := []m.TileID{}
	step := 0
	for rows.Next() {
		var x, y, z int

		rows.Scan(&x, &y, &z)
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x), int64(y), uint64(z)}
		mymap = append(mymap, tileid)
		step += 1
	}
	return mymap
}
