package mbutil

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"log"
	//"os"
	//"math/rand"
	"encoding/json"
	"reflect"
	"sync"
)

// vector layer json
type Vector_Layer struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Minzoom     int               `json:"minzoom"`
	Maxzoom     int               `json:"maxzoom"`
	Fields      map[string]string `json:"fields"`
}

type Vector_Layers struct {
	Vector_Layers []Vector_Layer `json:"vector_layers"`
}

// mbtiles struct
type Mbtiles struct {
	Tx        *sql.Tx    // Tx
	Stmt      *sql.Stmt  // stmt
	Mutex     sync.Mutex // mutex
	NewBool   bool       // a new bool for whether or not to add to existing db
	LayerName string     // the layer name currently commiting to
	Geobuf    bool       // a bool for whether or a byte array is a geobuf
	Total     int        // the total number of tiles iterated through
	Old_Total int        // the oldtotal
	FileName  string
	MinZoom   int
	MaxZoom   int
	Gzipped   bool
}

// configuration structure
type Config struct {
	FileName        string                 // filename of db
	Description     string                 // description of layer
	LayerName       string                 // the first layername to be added
	LayerProperties map[string]interface{} // an example of the fields or properties
	MinZoom         int                    // maxzoom
	MaxZoom         int                    // minzoom
	Geobuf          bool                   // a geobuf bool
}

// reflects a tile value back and stuff
func ReflectFields(mymap map[string]interface{}) map[string]string {
	newmap := map[string]string{}
	for k, v := range mymap {
		vv := reflect.ValueOf(v)
		kd := vv.Kind()
		if (reflect.Float64 == kd) || (reflect.Float32 == kd) {
			//fmt.Print(v, "float", k)
			newmap[k] = "Number"
			//hash = Hash_Tv(tv)
		} else if (reflect.Int == kd) || (reflect.Int8 == kd) || (reflect.Int16 == kd) || (reflect.Int32 == kd) || (reflect.Int64 == kd) || (reflect.Uint8 == kd) || (reflect.Uint16 == kd) || (reflect.Uint32 == kd) || (reflect.Uint64 == kd) {
			//fmt.Print(v, "int", k)
			newmap[k] = "Number"
			//hash = Hash_Tv(tv)
		} else if reflect.String == kd {
			//fmt.Print(v, "str", k)
			newmap[k] = "String"
			//hash = Hash_Tv(tv)

		} else {
			fmt.Print(k, v, "\n")
		}
	}
	return newmap
}

// returns the string of the json meta data
func MakeJsonMeta(config Config, json_meta string) string {
	// v vector_layers
	var vector_layers Vector_Layers
	_ = json.Unmarshal([]byte(json_meta), &vector_layers)

	layer := Vector_Layer{ID: config.LayerName, Description: "", Minzoom: config.MinZoom, Maxzoom: config.MaxZoom}

	fields := ReflectFields(config.LayerProperties)
	layer.Fields = fields

	vector_layers.Vector_Layers = append(vector_layers.Vector_Layers, layer)

	b, _ := json.Marshal(vector_layers)
	return string(b)
}

// creates a mbtiles file
func CreateDB(config Config) (Mbtiles, error) {
	// linting maxzoom
	if config.MaxZoom == 0 {
		config.MaxZoom = 20
	}

	db, err := sql.Open("sqlite3", config.FileName)
	if err != nil {
		return Mbtiles{}, err
	}

	// creating tiles table
	sqlStmt := `
	CREATE TABLE tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return Mbtiles{}, err
	}

	// creating metadata table
	sqlStmt = `
	CREATE TABLE metadata (name text, value text);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return Mbtiles{}, err
	}

	// getting metadata and inserting into table
	values := [][]string{{"name", config.FileName},
		{"type", "overlay"},
		{"version", "2"},
		{"description", config.Description},
		{"format", "pbf"},
		{"json", MakeJsonMeta(config, "")},
	}

	// doing the transaction for meta data
	tx, err := db.Begin()
	if err != nil {
		return Mbtiles{}, err
	}

	// metadata stmt
	stmt, err := tx.Prepare("insert into metadata(value, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	// inserting each metadata value
	for _, i := range values {
		_, err = stmt.Exec(i[1], i[0])
		if err != nil {
			return Mbtiles{}, err
		}
	}
	tx.Commit()

	// starting the transaction for adding tiles
	tx, err = db.Begin()
	if err != nil {
		return Mbtiles{}, err
	}

	// creaitng stmt for tiles
	stmt, err = tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		return Mbtiles{}, err
	}
	var mutex sync.Mutex
	mb := Mbtiles{Tx: tx,
		Stmt:      stmt,
		Mutex:     mutex,
		NewBool:   true,
		Geobuf:    config.Geobuf,
		Old_Total: -1,
		FileName:  config.FileName,
		MinZoom:   config.MinZoom,
		MaxZoom:   config.MaxZoom,
		LayerName: config.LayerName,
	}
	return mb, nil
}

// updates an mbtiles file
func UpdateDB(config Config) (Mbtiles, error) {
	// linting maxzoom
	if config.MaxZoom == 0 {
		config.MaxZoom = 20
	}

	db, err := sql.Open("sqlite3", config.FileName)
	if err != nil {
		return Mbtiles{}, err
	}

	// selecting metadata
	sqlStmt := `
	select value from metadata where name = "json";
	`
	var jsonstring string
	err = db.QueryRow(sqlStmt).Scan(&jsonstring)

	if err != nil {
		return Mbtiles{}, err
	}

	// updating metadata
	tx, err := db.Begin()
	if err != nil {
		return Mbtiles{}, err
	}

	stmt, err := db.Prepare("update metadata set value=? where name=?")
	if err != nil {
		return Mbtiles{}, err
	}

	_, err = stmt.Exec(MakeJsonMeta(config, jsonstring), "json")
	if err != nil {
		return Mbtiles{}, err
	}
	tx.Commit()

	// starting the transaction for adding tiles
	tx, err = db.Begin()
	if err != nil {
		return Mbtiles{}, err
	}

	// creaitng stmt for tiles
	stmt, err = tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		return Mbtiles{}, err
	}
	var mutex sync.Mutex
	mb := Mbtiles{Tx: tx,
		Stmt:      stmt,
		Mutex:     mutex,
		NewBool:   false,
		Geobuf:    config.Geobuf,
		Old_Total: -1,
		FileName:  config.FileName,
		MinZoom:   config.MinZoom,
		MaxZoom:   config.MaxZoom,
		LayerName: config.LayerName,
	}
	mb.CheckGZip()
	return mb, nil
}

// adds a single tile to sqlite db
func (mbtiles *Mbtiles) AddTile(k m.TileID, bytes []byte) error {
	k.Y = (1 << uint64(k.Z)) - 1 - k.Y
	if mbtiles.NewBool == false {
		mbtiles.Mutex.Lock()
		var data []byte
		query := fmt.Sprintf("select tile_data from tiles where zoom_level = %d and tile_column = %d and tile_row = %d", k.Z, k.X, k.Y)
		err := mbtiles.Tx.QueryRow(query).Scan(&data)
		mbtiles.Mutex.Unlock()
		if len(data) > 0 {
			bytes = append(bytes, data...)
			mbtiles.Mutex.Lock()
			_, err = mbtiles.Tx.Exec(`update tiles set tile_data = ? where zoom_level = ? and tile_column = ? and tile_row = ?`, bytes, k.Z, k.X, k.Y)
			if err != nil {
				return err
			}
			mbtiles.Mutex.Unlock()
		} else {
			mbtiles.Mutex.Lock()
			_, err = mbtiles.Stmt.Exec(k.Z, k.X, k.Y, bytes)
			if err != nil {
				return err
			}
			mbtiles.Mutex.Unlock()
		}
	} else {
		mbtiles.Mutex.Lock()
		_, err := mbtiles.Stmt.Exec(k.Z, k.X, k.Y, bytes)
		if err != nil {
			return err
		}
		mbtiles.Mutex.Unlock()
	}
	return nil
}

// replaces a single tile to sqlite db
func (mbtiles *Mbtiles) ReplaceTile(k m.TileID, bytes []byte) error {
	k.Y = (1 << uint64(k.Z)) - 1 - k.Y
	mbtiles.Mutex.Lock()
	_, err := mbtiles.Tx.Exec(`update tiles set tile_data = ? where zoom_level = ? and tile_column = ? and tile_row = ?`, bytes, k.Z, k.X, k.Y)
	if err != nil {
		return err
	}
	mbtiles.Mutex.Unlock()
	return nil
}

// getting min and maxzoom from metadata
func (mbtiles *Mbtiles) GetMinMaxZoom() (int, int) {
	// selecting metadata
	sqlStmt := `
	select value from metadata where name = "json";
	`
	var jsonstring string
	err := mbtiles.Tx.QueryRow(sqlStmt).Scan(&jsonstring)

	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}

	// unmarshalling json metadata
	var vector_layers Vector_Layers
	_ = json.Unmarshal([]byte(jsonstring), &vector_layers)
	var minzoom, maxzoom int
	minzoom = 20
	maxzoom = 0
	for _, layer := range vector_layers.Vector_Layers {
		tmpmin, tmpmax := layer.Minzoom, layer.Maxzoom
		if tmpmin < minzoom {
			minzoom = tmpmin
		}
		if tmpmax > maxzoom {
			maxzoom = tmpmax
		}
	}

	// setting defaults if given
	if minzoom == 20 {
		minzoom = 0
	}
	if maxzoom == 0 {
		maxzoom = 20
	}
	return minzoom, maxzoom
}

// commiting and updating index
func (mbtiles *Mbtiles) Commit() error {
	sqlStmt := `
	CREATE UNIQUE INDEX IF NOT EXISTS tile_index on tiles (zoom_level, tile_column, tile_row)
	`
	_, err := mbtiles.Tx.Exec(sqlStmt)
	if err != nil {
		fmt.Println(err)
	}

	err = mbtiles.Tx.Commit()
	// reading db again
	db, err := sql.Open("sqlite3", mbtiles.FileName)
	if err != nil {
		return err
	}

	_,err = db.Exec("VACUUM;")
	if err != nil {
		return err
	}

	// starting the transaction for adding tiles
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// creaitng stmt for tiles
	stmt, err := tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	mbtiles.Stmt = stmt
	mbtiles.Tx = tx
	//mbtiles.Stmt.Close()
	return err
}


// adds a single tile to sqlite db
func (mbtiles *Mbtiles) CheckGZip() {
	if GetFileSize(mbtiles.FileName) > 0 {
		tile,_ := mbtiles.GetOneTile()
		data,err := mbtiles.QueryRaw(tile)
		if err != nil {
			fmt.Println(err)
		}
		if len(data) > 0 {
			mbtiles.Gzipped = (data[0] == 0x1f) && (data[1] == 0x8b)
		} else {
			fmt.Println("first tile size equal to zero")
		}
	} else {
		fmt.Println("No file found skippign CheckGZip")
	}
}