package main

import (
	"log"
	"os"
	"strings"

	"github.com/Fitzgeraldjc/GoTorrent/magnets"
	"github.com/Fitzgeraldjc/GoTorrent/torrentfile"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: GoTorrent <torrent-file-or-magnet-link> <output-path>")
	}

	input := os.Args[1]
	outPath := os.Args[2]

	if strings.HasPrefix(input, "magnet:") {
		magnet, err := magnets.Parse(input)
		if err != nil {
			log.Fatalf("Failed to parse magnet link: %v", err)
		}

		err = magnet.DownloadToFile(outPath)
		if err != nil {
			log.Fatalf("Failed to download from magnet link: %v", err)
		}
	} else {
		tf, err := torrentfile.Open(input)
		if err != nil {
			log.Fatalf("Failed to open torrent file: %v", err)
		}

		err = tf.DownloadToFile(outPath)
		if err != nil {
			log.Fatalf("Failed to download torrent: %v", err)
		}
	}

	log.Println("Download completed successfully!")
}
