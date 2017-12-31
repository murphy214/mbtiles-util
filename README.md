# mbtiles-util

# Summary 

This project is a few commonly used patterns I use when creating or editing mbtiles files, hopefully by having it all in one place I can simplify my code base a little bit and add abstractions for converting between other file formats. The create / update feature for the mbtiles file format is layer transactional in nature. Meaning every transaction open and commit is for the purposes of creating a new mbtiles and adding a layer or adding a layer to an existing mbtiles file. This pattern works quite well and works for most use cases, and is ideal because no protobuf serialization is necessary everything here is done in raw bytes which is much less cpu and memory intensive. 

The Add_Tile function is wrapped in a mutex as to not offend the sqlite implmentation in go that uses cgo, and single threaded therefore there is no need to worry about adding tiles in go functions becoming a problem. Layer properties could literally be considered or assumed to be any properties value from a geojson feature in your dataset.

# Example

The following example shows me creating a test.mbtiles file with the layer "Test". Then updating the test.mbtiles file with "Test2" layer, finally showing how to use the query feature as well. Nothing to crazy. 

```go
package main 

import (
	util "./mbtiles_util"
	"fmt"
	m "github.com/murphy214/mercantile"

)


func main() {
	// configuration for creating and updating mbtiles	
	config := util.Config{LayerName:"Test",
				FileName:"test.mbtiles",
				LayerProperties:map[string]interface{}{"shit":10},
			}

	// layer creation / update instance created here
	mbtile := util.Create_DB(config)

	// adding a single tile at 0,0,0
	mbtile.Add_Tile(m.TileID{0,0,0},[]byte{10,10,10})
	
	// commiting the initial layer
	mbtile.Commit()

	// configuration for creating and updating mbtiles	
	config_update := util.Config{LayerName:"Test2",
				FileName:"test.mbtiles",
				LayerProperties:map[string]interface{}{"shit":10},
			}

	// layer creation / update instance created here
	mbtile_update := util.Update_DB(config_update)

	// adding a single tile at 0,0,0
	mbtile_update.Add_Tile(m.TileID{0,0,0},[]byte{10,10,10})
	
	// commiting the updated layer
	mbtile_update.Commit()

	// loading the mbtiles file as a read instance for querying
	mbtile = util.Read_Mbtiles("test.mbtiles")
	fmt.Println(mbtile.Query(m.TileID{0,0,0}))

}
```
