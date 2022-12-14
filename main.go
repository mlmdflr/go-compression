package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	h bool
	c string
)

type Config struct {
	ZipPath      string   `json:"zipPath"`
	SourceFolder string   `json:"sourceFolder"`
	Src          []string `json:"src"`
}

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&c, "c", "config.json", "Configuration file path")
	flag.StringVar(&c, "config", "config.json", "Configuration file path")
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
	}
	file, err := os.ReadFile(c)
	if err != nil {
		log.Fatalf("Some error occured while reading file. Error: %s", err)
	}
	cs := &Config{}
	err = json.Unmarshal(file, cs)
	if err != nil {
		log.Fatalf("Error occured during unmarshaling. Error: %s", err.Error())
	}
	_, err = compression(cs.ZipPath, cs.SourceFolder, cs.Src...)
	if err != nil {
		log.Printf("Exception during compression. Error: %s", err.Error())
	}
}

func compression(zipPath string, sourceFolder string, src ...string) (string, error) {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return zipPath, err
	}
	defer func(zipFile *os.File) {
		err := zipFile.Close()
		if err != nil {
			log.Printf("zipFile Close Error: %s", err.Error())
		}
	}(zipFile)

	w := zip.NewWriter(zipFile)

	defer func(w *zip.Writer) {
		err := w.Close()
		if err != nil {
			log.Printf("zipWriter Close Error: %s", err.Error())
		}
	}(w)

	for _, s := range src {
		err := func(path string) error {
			path = filepath.Clean(path)
			path = strings.Trim(path, string(filepath.Separator))

			sFile, err := os.Open(filepath.Join(sourceFolder, path))
			if err != nil {
				return err
			}
			defer func(sFile *os.File) {
				_ = sFile.Close()
			}(sFile)

			sFileInfo, err := sFile.Stat()
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(sFileInfo)
			if err != nil {
				return err
			}

			//启用压缩
			header.Method = zip.Deflate
			//保持目录
			header.Name = path

			headerWriter, err := w.CreateHeader(header)
			_, err = io.Copy(headerWriter, sFile)
			if err != nil {
				return err
			}
			return nil
		}(s)
		if err != nil {
			return zipPath, err
		}
	}
	return zipPath, nil
}
