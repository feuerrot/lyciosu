package storage

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/feuerrot/lyciosu/camera"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
	bucket string
	files  map[string]*camera.Image
}

type Config struct {
	Username string
	Password string
	Hostname string
	Bucket   string
	Secure   bool
}

func NewStorage(ctx context.Context, config *Config) (*Storage, error) {
	stor := Storage{}
	stor.bucket = config.Bucket

	mc, err := minio.New(config.Hostname, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Username, config.Password, ""),
		Secure: config.Secure,
	})
	if err != nil {
		return &stor, fmt.Errorf("Kann Storageclient nicht initialisieren: %v", err)
	}
	stor.client = mc

	exists, err := stor.client.BucketExists(ctx, stor.bucket)
	if err != nil {
		return &stor, fmt.Errorf("Konnte Existenz von \"%s\" nicht prüfen: %v", stor.bucket, err)
	}
	if !exists {
		return &stor, fmt.Errorf("Bucket \"%s\" existiert nicht", stor.bucket)
	}

	return &stor, nil
}

func (s *Storage) UploadQueue(ctx context.Context, images chan *camera.Image) (chan *camera.Image, chan error) {
	out := make(chan *camera.Image)
	errchan := make(chan error)
	go func() {
		var wg sync.WaitGroup
		for image := range images {
			wg.Add(1)
			image := image
			go func() {
				if err := s.Upload(ctx, image); err != nil {
					errchan <- fmt.Errorf("Fehler beim Upload von %s: %v", image.Name, err)
				}
				out <- image
				wg.Done()
			}()
		}
		wg.Wait()
		close(errchan)
		close(out)
	}()

	return out, errchan
}

func (s *Storage) Upload(ctx context.Context, image *camera.Image) error {
	ts_start := time.Now()
	res, err := s.client.PutObject(ctx, s.bucket, image.Name, image, image.Size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	ts_end := time.Now()
	image.Close()
	if err != nil {
		log.Printf("Fehler beim Upload: %v", err)
	}
	if res.Size != image.Size {
		log.Printf("%s: Unterschied Größe und Upload: %d vs. %d", image.Name, image.Size, res.Size)
	}

	image.Duration("upload", ts_end.Sub(ts_start))

	return nil
}
