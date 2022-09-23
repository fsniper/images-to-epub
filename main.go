package main

import (
	"flag"
	"fmt"
	"github.com/bmaupin/go-epub"
	"github.com/disintegration/imaging"
	"io/fs"
	"log"
	"path/filepath"
  "image"
  "os"
)

func is_landscape(img image.Image) bool {
  var max_diff_ratio_in_y_to_x float32 =  1.1
  bounds := img.Bounds()
  width := float32(bounds.Max.X - bounds.Min.X)
  height := float32(bounds.Max.Y - bounds.Min.Y) 
  height_adjusted := float32(height) * max_diff_ratio_in_y_to_x

  return (width > height_adjusted)
}

func convert_to_portrait_mode(image string) (string, bool) {
	src, err := imaging.Open(image)
  
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

  if is_landscape(src) {
    portrait_image := imaging.Rotate90(src)
    tmp, err := os.CreateTemp("", fmt.Sprintf("*-%s", filepath.Base(image)))
	  if err != nil {
		  log.Fatalf("failed to create tmp file: %v", err)
	  }
    
    log.Printf("Saving portrait mode version to %s", tmp.Name())
    err = imaging.Save(portrait_image, tmp.Name())
	  if err != nil {
		  log.Fatalf("failed to save image: %v", err)
	  }
	  return tmp.Name(), true
  }

	return image, false
}

func main() {

	var chapters []string

	var path = flag.String("path", ".", "path of images")
	var title = flag.String("title", "My Epub Title", "title of the epub")
	var convert_portrait = flag.Bool("to_portrait", false, "Convert all images to portrait")
	var output = flag.String("output", "my-epub.epub", "output file path")
	flag.Parse()

	iterateImages := func(path string, direntry fs.DirEntry, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}

		if !direntry.IsDir() {
			fmt.Println("Adding image", path)
			chapters = append(chapters, path)
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	}

	err := filepath.WalkDir(*path, iterateImages)
	if err != nil {
		log.Print(err)
	}

	e := epub.NewEpub(*title)

  var rotated bool
  var image string

	for _, chapter := range chapters {

    if *convert_portrait {
		  chapter, rotated = convert_to_portrait_mode(chapter)
    }
		image, err = e.AddImage(chapter, "")

		section := fmt.Sprintf("<img src=\"%s\" alt=\"%s\"/>", image, chapter)
		log.Printf(section)
		e.AddSection(section, chapter, "", "")
    if rotated {
      defer os.Remove(chapter)
    }
	}

	err = e.Write(*output)
	if err != nil {
		log.Fatal(err)
	}
}
