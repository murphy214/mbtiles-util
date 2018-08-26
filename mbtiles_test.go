package mbutil

import (
	m "github.com/murphy214/mercantile"
	"os"
	"testing"
)

var config = Config{
	FileName:        "a.mbtiles",
	LayerName:       "test",
	MinZoom:         0,
	MaxZoom:         20,
	LayerProperties: map[string]interface{}{},
}

var config2 = Config{
	FileName:        "a.mbtiles",
	LayerName:       "test2",
	MinZoom:         0,
	MaxZoom:         20,
	LayerProperties: map[string]interface{}{},
}

func TestCreateDBAddTile(t *testing.T) {
	os.Remove("a.mbtiles")
	db, err := CreateDB(config)
	if err != nil {
		t.Errorf("Error Test CreateDB\n", db.FileName)
	}
	err = db.AddTile(m.TileID{0, 0, 0}, []byte{29})
	if err != nil {
		t.Errorf("Error Test CreateDB in Add Tile\n")
	}

	err = db.Commit()
	if err != nil {
		t.Errorf("Error Test CreateDB in Commit\n")
	}

	val, err := db.Query(m.TileID{0, 0, 0})
	if err != nil && val[0] == byte(29) {
		t.Errorf("Error Test CreateDB in Query\n")
	}
}

func TestUpdateDBAddTile(t *testing.T) {
	db, err := UpdateDB(config)
	if err != nil {
		t.Errorf("Error Test CreateDB\n", db.FileName)
	}
	err = db.AddTile(m.TileID{0, 0, 0}, []byte{29})
	if err != nil {
		t.Errorf("Error Test CreateDB in Add Tile\n")
	}

	err = db.Commit()
	if err != nil {
		t.Errorf("Error Test CreateDB in Commit\n")
	}

	val, err := db.Query(m.TileID{0, 0, 0})
	if err != nil && val[1] == byte(29) {
		t.Errorf("Error Test CreateDB in Query\n")
	}
	os.Remove("a.mbtiles")
}
