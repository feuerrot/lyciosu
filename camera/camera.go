package camera

import (
	"context"
	"fmt"
	"log"

	"github.com/frizinak/gphoto2go"
)

type Camera struct {
	gphoto2go.Camera
	ctx context.Context
}

func New(ctx context.Context) (*Camera, error) {
	camera := new(Camera)
	if err := camera.Init(); err != nil {
		return nil, fmt.Errorf("Kann Kamera nicht initialisieren: %v", err)
	}
	camera.ctx = ctx

	model, _ := camera.Model()
	log.Printf("Kameramodell: %s", model)

	go func() {
		<-camera.ctx.Done()
		if err := camera.Exit(); err != nil {
			log.Printf("Kann Kamera nicht schließen: %v", err)
		}
	}()

	return camera, nil

}

func (c *Camera) Images() ([]*Image, error) {
	images := []*Image{}
	folders, err := c.RListFolders("/")
	if err != nil {
		return images, fmt.Errorf("Kann Verzeichnisse nicht auflisten: %v", err)
	}

	for _, folder := range folders {
		files, err := c.ListFiles(folder)
		if err != nil {
			return images, fmt.Errorf("Kann \"%s\" nicht lesen: %v", folder, err)
		}
		for _, file := range files {
			img, err := c.NewImage(folder, file)
			if err != nil {
				return images, fmt.Errorf("Kann Bild nicht erstellen: %v", err)
			}
			images = append(images, img)
		}
	}

	return images, nil
}

func (c *Camera) LoadImages(ctx context.Context, images []*Image) chan *Image {
	out := make(chan *Image)
	go func() {
		for _, img := range images {
			select {
			case <-ctx.Done():
				close(out)
				return
			default:
			}
			img.Open(c)
			out <- img
		}
		close(out)
	}()

	return out
}
