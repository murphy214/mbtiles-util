package mbutil

import (
	"github.com/murphy214/vector-tile-go"
	"fmt"
	"bytes"
	"encoding/json"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)


func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func (mbtiles *Mbtiles) UpdateMetaDataKV(key,value string) bool {
	stmt, err := mbtiles.Tx.Prepare("update metadata set value = ? where name = ?;")

	result, err := stmt.Exec(string(value),string(key))
	effected,_  := result.RowsAffected()
	if effected == 0 {
		stmt, err := mbtiles.Tx.Prepare("insert into metadata(name, value) values(?, ?)")
		if err != nil {
			fmt.Println(err)
		}
		_, err = stmt.Exec(key,value)
	}

	if err != nil {
		fmt.Println(err)
	}

	err = mbtiles.Tx.Commit()
	if err != nil {
		fmt.Println(err)
	}
	db, err := sql.Open("sqlite3", mbtiles.FileName)
	if err != nil {
		fmt.Println(err)
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
	}

	// creaitng stmt for tiles
	stmt, err = tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
	}
	mbtiles.Stmt = stmt
	mbtiles.Tx = tx

	return err == nil

}

func UpdateMetaDataKV(mbfilename,key,value string) bool {
	mbtiles,err := ReadMbtiles(mbfilename)
	if err != nil {
		fmt.Println(err)
	}
	return mbtiles.UpdateMetaDataKV(key,value)
}

// updates the underlyig metadata of a json field
func UpdateMetaDataJSON(mbfilename string) {
	mbtiles,err := ReadMbtiles(mbfilename)
	if err != nil {
		fmt.Println(err)
	}
	mymap := map[string]Vector_Layer{}
	tiles := mbtiles.GetAllTiles()
	sizetiles := len(tiles)
	for pos,tile := range tiles {
		myzoom := int(tile.Z)
		bs,err := mbtiles.Query(tile)
		if err != nil {
			fmt.Println(err)
		}
		tile,err := vt.NewTile(bs)
		if err != nil {
			fmt.Println(err)
		}
		for _,layername := range tile.Layers {
			vector_layer,boolval := mymap[layername]
			if !boolval {
				vector_layer = Vector_Layer{ID:layername,Minzoom:int(myzoom),Maxzoom:int(myzoom),Fields:map[string]string{}}
				mymap[layername] = vector_layer
			} else {
				if vector_layer.Minzoom > myzoom {
					vector_layer.Minzoom = myzoom
				}

				if vector_layer.Maxzoom < myzoom {
					vector_layer.Maxzoom = myzoom
				}
				mymap[layername] = vector_layer
			}
		}
		if pos%1000==0 {
			fmt.Printf("\r[%d/%d]",pos,sizetiles)			
		}
	}
	layers := []Vector_Layer{}
	for _,v := range mymap {
		layers = append(layers,v)
	}
	vlayers := Vector_Layers{Vector_Layers:layers}
	b, _ := json.Marshal(vlayers)
	b,_ = prettyprint(b)
	fmt.Println(string(b))

	// metadata stmt
	stmt, err := mbtiles.Tx.Prepare("update metadata set value = ? where name = 'json';")

	_, err = stmt.Exec(string(b))
	if err != nil {
		fmt.Println(err)
	}

	err = mbtiles.Tx.Commit()
	if err != nil {
		fmt.Println(err)
	}
}

// updates and removes given layers
func RemoveLayers(mbfilename string,layernames []string) {
	mbtiles,err := ReadMbtiles(mbfilename)
	if err != nil {
		fmt.Println(err)
	}

	for _,tile := range mbtiles.GetAllTiles() {
		bs,err := mbtiles.Query(tile)
		if err != nil {
			fmt.Println(err)
		}
		tilee,err := vt.NewTile(bs)
		if err != nil {
			fmt.Println(err)
		}
		bss := tilee.DeleteLayers(layernames)
		err = mbtiles.ReplaceTile(tile,bss)
		if err != nil {
			fmt.Println(err)
		}
	}
	
	err = mbtiles.Commit()
	if err != nil {
		fmt.Println(err)
	}
	UpdateMetaDataJSON(mbfilename)

}

// gzips all layers
func GZipAll(mbfilename string) {
	mbtiles,err := ReadMbtiles(mbfilename)
	if err != nil {
		fmt.Println(err)
	}
	mytiles := mbtiles.GetAllTiles()
	for pos,tile := range mytiles {
		bs,err := mbtiles.Query(tile)
		if err != nil {
			fmt.Println(err)
		}
		bs = GZipWrite(bs)
		err = mbtiles.ReplaceTile(tile,bs)
		if err != nil {
			fmt.Println(err)
		}
		if pos%1000 == 0 {
			fmt.Printf("\r[%d/%d] GZipping all tiles...",pos,len(mytiles))
		}
	}

	err = mbtiles.Commit()
	if err != nil {
		fmt.Println(err)
	}
}


func FixViewToTable(filename string) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		fmt.Println(err)

	}

	_, err = db.Exec(`CREATE TABLE tiles2 AS select * from tiles;`)
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Exec(`DROP VIEW tiles;`)
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Exec("ALTER TABLE `tiles2` RENAME TO `tiles`;")
	if err != nil {
		fmt.Println(err)
	}

	mbtiles,err := ReadMbtiles(filename)
	if err != nil {
		fmt.Println(err)
	}
	mbtiles.Commit()
}