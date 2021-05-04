package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/feuerrot/lyciosu/camera"
	"github.com/feuerrot/lyciosu/storage"
	"github.com/feuerrot/lyciosu/util"
	"github.com/spf13/viper"
)

func readConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Kann Konfiguration nicht lesen: %v", err)
	}

	return nil
}

func initStorage(ctx context.Context) (*storage.Storage, error) {
	config := &storage.Config{}
	if err := viper.UnmarshalKey("storage", config); err != nil {
		return nil, fmt.Errorf("Konfiguration kann nicht unmarshalled werden: %v", err)
	}

	storage, err := storage.NewStorage(ctx, config)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func initContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt)

	go func() {
		select {
		case <-sigchan:
			log.Printf("received SIGINT, shutting down")
			cancel()
			time.Sleep(time.Second * 2)
			os.Exit(1)
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}

func main() {
	log.Printf("Initialisiere Kontext")
	ctx, cancel := initContext()

	log.Printf("Lese Konfiguration")
	if err := readConfig(); err != nil {
		log.Fatalf("Konfiguration kann nicht gelesen werden: %v", err)
	}

	log.Printf("Initialisiere Storage")
	storage, err := initStorage(ctx)
	if err != nil {
		log.Fatalf("Kann Storage nicht initialisieren: %v", err)
	}

	log.Printf("Initialisiere Kamera")
	camera, err := camera.New(ctx)
	if err != nil {
		log.Fatalf("Kann Kamera nicht öffnen: %v", err)
	}

	ts_load := time.Now()
	log.Printf("Lade Bilder")
	images, err := camera.Images()
	if err != nil {
		log.Fatalf("Kann Bilder nicht laden: %v", err)
	}

	queue := camera.LoadImages(ctx, images)
	uploaded, errchan := storage.UploadQueue(ctx, queue)
	go func() {
		for {
			select {
			case err := <-errchan:
				if err != nil {
					log.Printf("%v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	var size int64
	for img := range uploaded {
		select {
		case <-ctx.Done():
			break
		default:
		}
		size += img.Size
		log.Printf("%s", img)
	}
	ts_done := time.Now()

	log.Printf("Summe: %s", util.FSizeDurationSpeed(size, ts_done.Sub(ts_load)))
	cancel()
}
