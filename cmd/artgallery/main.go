//go:build artgallery

// Artgallery exports the game's real ASCII renderers for the PNG contact sheets.
package main

import (
	"flag"
	"github.com/rene-timmermans/ttymagotchi/internal/app"
	"log"
)

func main() {
	output := flag.String("out", "art-gallery.json", "Output JSON path")
	flag.Parse()
	if err := app.ExportArtGallery(*output); err != nil {
		log.Fatal(err)
	}
}
