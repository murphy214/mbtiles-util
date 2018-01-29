package mbutil


import (
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"log"
	//"os"
	"sync"
	"github.com/paulmach/go.geojson"
	g "github.com/murphy214/geobuf"

)

// given a filename gets a filename structure to read 
func Read_Mbtiles(filename string) Mbtiles {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		fmt.Println(err)
	}
	// starting the transaction for adding tiles
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	var mutex sync.Mutex
	mb := Mbtiles{Tx:tx,
				Mutex:mutex,
				NewBool:false,
				Old_Total:-1,
				FileName:filename}
	minzoom,maxzoom := mb.Get_Min_Max_Zoom()
	mb.MinZoom = minzoom
	mb.MaxZoom = maxzoom
	return mb

}

// queries a given tileid
func (mbtiles *Mbtiles) Query(k m.TileID) []byte {
	k.Y = (1 << uint64(k.Z)) - 1 - k.Y 
	mbtiles.Mutex.Lock()
	var data []byte
	err := mbtiles.Tx.QueryRow("select tile_data from tiles where zoom_level = ? and tile_column = ? and tile_row = ?",k.Z,k.X,k.Y).Scan(&data)
	if err != nil {
		fmt.Println(err)
	}
	mbtiles.Mutex.Unlock()
	return data
}	


// queries a given tileid
func (mbtiles *Mbtiles) Query_Features(k m.TileID) map[string][]*geojson.Feature {
	og := k
	k.Y = (1 << uint64(k.Z)) - 1 - k.Y 
	mbtiles.Mutex.Lock()
	var data []byte
	err := mbtiles.Tx.QueryRow("select tile_data from tiles where zoom_level = ? and tile_column = ? and tile_row = ?",k.Z,k.X,k.Y).Scan(&data)
	if err != nil {
		//fmt.Println(err)
	}
	mbtiles.Mutex.Unlock()
	if mbtiles.Geobuf == false {
		return Convert_Vt_Bytes(data,og)
	} else {
		return map[string][]*geojson.Feature{"geobuf":g.Read_FeatureCollection(data)}
	}
}

// function for getting the next value
func (mbtiles *Mbtiles) Next() bool {
	return (mbtiles.Old_Total == mbtiles.Total) == false
}


// map query 
func (mbtiles *Mbtiles) Map_Tiles_Get(limit int) map[m.TileID][]byte {
	rows,err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles limit ?,?",mbtiles.Total,limit)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid 
	mymap := map[m.TileID][]byte{}
	step := 0
	for rows.Next() {
		var x,y,z int
		var data []byte

		rows.Scan(&x,&y,&z,&data)
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x),int64(y),uint64(z)}
		mymap[tileid] = data
		step += 1
	}
	mbtiles.Old_Total = mbtiles.Total
	mbtiles.Total += step

	return mymap
}


// map query 
func (mbtiles *Mbtiles) Chunk_Tiles(limit int) map[m.TileID][]byte {
	rows,err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles limit ?,?",mbtiles.Total,limit)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid 
	mymap := map[m.TileID][]byte{}
	step := 0
	for rows.Next() {
		var x,y,z int
		var data []byte

		rows.Scan(&x,&y,&z,&data)
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x),int64(y),uint64(z)}
		mymap[tileid] = data
		step += 1
	}
	mbtiles.Old_Total = mbtiles.Total
	mbtiles.Total += step

	return mymap
}

// map query 
func (mbtiles *Mbtiles) Chunk_Tiles_Zoom(limit int,zoom int) map[m.TileID][]byte {
	rows,err := mbtiles.Tx.Query("SELECT tile_column,tile_row,zoom_level,tile_data FROM tiles where zoom_level = ? limit ?,?",zoom,mbtiles.Total,limit)
	if err != nil {
		fmt.Println(err)
	}

	// mapping to a specific tileid 
	mymap := map[m.TileID][]byte{}
	step := 0
	for rows.Next() {
		var x,y,z int
		var data []byte

		rows.Scan(&x,&y,&z,&data)
		y = (1 << uint64(z)) - y - 1
		tileid := m.TileID{int64(x),int64(y),uint64(z)}
		mymap[tileid] = data
		step += 1
	}
	mbtiles.Old_Total = mbtiles.Total
	mbtiles.Total += step

	return mymap
}

