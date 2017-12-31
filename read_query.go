package mbutil


import (
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"log"
	//"os"
	"sync"
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
	return  Mbtiles{Tx:tx,Mutex:mutex,NewBool:false}

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