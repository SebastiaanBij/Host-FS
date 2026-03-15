package main

import (
	"fmt"
	"host-fs/src/lib"
	"log"
	"os"
	"path"

	"github.com/docker/go-plugins-helpers/volume"
)

func main() {
	driver, err := lib.NewHostFSDriver(path.Join(os.Getenv("HOST_DIR"), os.Getenv("STATE_DIR")))
	if err != nil {
		log.Fatalf("could not instantiate Host FS Driver, due to the following error: %s", err.Error())
	}

	handler := volume.NewHandler(driver)
	fmt.Println(handler.ServeUnix("host-fs", 0))
}
