package main

import (
	"io/fs"
	"log"
	"os"

	"github.com/AndreRenaud/spectacle/ui"
)

func main() {
	fsys, err := fs.Sub(os.DirFS("."), ".")
	if err != nil {
		log.Fatal(err)
	}
	if err := ui.Run(fsys.(fs.ReadDirFS)); err != nil {
		log.Fatal(err)
	}
}
