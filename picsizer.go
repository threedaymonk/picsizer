package main

import (
	"code.google.com/p/gcfg"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/mcuadros/go-defaults"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Conversion struct {
	Operation string
	Width     int
	Height    int
}

type Config struct {
	Server struct {
		Address  string `default:"localhost"`
		Port     int    `default:8080`
		BaseDir  string
		CacheDir string `default:"./cache"`
	}
	Format map[string]*Conversion
}

var config Config

func main() {
	log.Printf("Setting GOMAXPROCS to %d", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())

	configFile := "picsizer.ini"

	defaults.SetDefaults(&config)
	gcfg.ReadFileInto(&config, configFile)

	log.Printf("%v", config.Format)

	listenOn := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)

	http.HandleFunc("/", handler)

	log.Printf("Listening on %s", listenOn)
	http.ListenAndServe(listenOn, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.URL.Path, "/", 3)
	slug := parts[1]
	path := parts[2]

	format, hasFormat := config.Format[slug]

	if !hasFormat {
		log.Printf("Unknown format %s", slug)
		http.ServeFile(w, r, "bogus")
		return
	}

	cachePath := filepath.Join(config.Server.CacheDir, slug, path)

	exists, _ := fileExists(cachePath)
	if !exists {
		originalPath := filepath.Join(config.Server.BaseDir, path)
		err := convertFile(originalPath, cachePath, format)
		if err != nil {
			log.Printf("Error: %s", err)
			// TODO: Generate stub image?
		}
	}

	log.Printf("Serving existing file %s", cachePath)
	http.ServeFile(w, r, cachePath)
}

func convertFile(src string, dest string, cnv *Conversion) error {
	log.Printf("Convert %s => %s via %s (%d, %d)", src, dest,
		cnv.Operation, cnv.Width, cnv.Height)

	img, err := imaging.Open(src)
	if err != nil {
		return err
	}

	switch cnv.Operation {
	case "copy":
		img = imaging.Clone(img)
	case "thumbnail":
		img = imaging.Thumbnail(img, cnv.Width, cnv.Height, imaging.CatmullRom)
	case "resize":
		img = imaging.Resize(img, cnv.Width, cnv.Height, imaging.Lanczos)
	case "fit":
		img = imaging.Fit(img, cnv.Width, cnv.Height, imaging.Lanczos)
	default:
		return fmt.Errorf("Unrecognised conversion operation: %s", cnv.Operation)
	}

	dir, _ := filepath.Split(dest)
	os.MkdirAll(dir, 0777)
	err = imaging.Save(img, dest)
	return err
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
