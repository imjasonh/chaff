package main

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const whiteoutPrefix = ".wh."

func main() {
	if len(os.Args) != 2 {
		log.Fatal("must provide one arg")
	}
	ref, err := name.ParseReference(os.Args[1])
	if err != nil {
		log.Fatalf("parsing reference: %v", err)
	}
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		log.Fatalf("fetching image: %v", err)
	}
	ls, err := img.Layers()
	if err != nil {
		log.Fatalf("getting layers: %v", err)
	}
	fileMap := map[string]bool{}
	var r report
	for i := len(ls) - 1; i >= 0; i-- {
		l := ls[i]

		rc, err := l.Uncompressed()
		if err != nil {
			log.Fatalf("getting layer %d: %v", i, err)
		}
		defer rc.Close()
		tr := tar.NewReader(rc)
		for {
			h, err := tr.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Fatalf("reading tar: %w", err)
			}

			basename := filepath.Base(h.Name)
			dirname := filepath.Dir(h.Name)
			tombstone := strings.HasPrefix(basename, whiteoutPrefix)
			if tombstone {
				basename = basename[len(whiteoutPrefix):]
			}

			// check if we have seen value before
			// if we're checking a directory, don't filepath.Join names
			var name string
			if h.Typeflag == tar.TypeDir {
				name = h.Name
			} else {
				name = filepath.Join(dirname, basename)
			}

			if _, ok := fileMap[name]; ok {
				continue
			}

			// check for a whited out parent directory
			if inWhiteoutDir(fileMap, name) {
				if h.Size > 0 {
					r.files = append(r.files, file{
						name: name,
						size: h.Size,
					})
					r.count++
					r.size += h.Size
				}
				continue
			}

			// mark file as handled. non-directory implicitly tombstones
			// any entries with a matching (or child) name
			fileMap[name] = tombstone || !(h.Typeflag == tar.TypeDir)
		}
	}

	fmt.Println("==== CHAFF REPORT ====")
	fmt.Println("- total chaff files:", r.count)
	fmt.Println("- total chaff size:", r.size)
	for _, f := range r.files {
		fmt.Printf("--- %s (%d bytes)\n", f.name, f.size)
	}
}

type report struct {
	size  int64
	count int
	files []file
}

type file struct {
	name string
	size int64
}

func inWhiteoutDir(fileMap map[string]bool, file string) bool {
	for {
		if file == "" {
			break
		}
		dirname := filepath.Dir(file)
		if file == dirname {
			break
		}
		if val, ok := fileMap[dirname]; ok && val {
			return true
		}
		file = dirname
	}
	return false
}
