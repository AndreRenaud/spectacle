package main

import (
	"io/fs"
	"log"
	"os"

	"github.com/AndreRenaud/spectacle/ui"
)

func main() {
	fsys := os.DirFS(".")
	if err := ui.Run(fsys.(fs.ReadDirFS)); err != nil {
		log.Fatal(err)
	}
}
