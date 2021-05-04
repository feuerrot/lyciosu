package camera

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/feuerrot/lyciosu/util"
)

const timefmt = "2006-01-02 15:04:05"

type Image struct {
	Name      string
	Path      string
	Size      int64
	Timestamp int64
	buffer    io.ReadCloser
	duration  map[string]time.Duration
}

func (c *Camera) NewImage(path, file string) (*Image, error) {
	img := Image{}

	info, err := c.Info(path, file)
	if err != nil {
		return &img, fmt.Errorf("Kann Information zu \"%s/%s\" nicht abrufen: %v", path, file, err)
	}

	img.Name = file
	img.Path = path
	img.Size = info.Size
	img.Timestamp = info.MTime
	img.duration = make(map[string]time.Duration)

	return &img, nil
}

func (i *Image) Duration(s string, t time.Duration) {
	if d, ok := i.duration[s]; ok {
		i.duration[s] += d + t
	} else {
		i.duration[s] = t
	}
}

func (i *Image) String() string {
	parts := []string{}
	parts = append(parts, fmt.Sprintf("%s/%s: %s", i.Path, i.Name, time.Unix(i.Timestamp, 0).Format(timefmt)))
	for time := range i.duration {
		parts = append(parts, fmt.Sprintf("\t%s:\t%s", time, util.FSizeDurationSpeed(i.Size, i.duration[time])))
	}

	return strings.Join(parts, "\n")
}

func (i *Image) Open(c *Camera) {
	ts_begin := time.Now()
	i.buffer = c.FileReader(i.Path, i.Name)
	i.Duration("open", time.Now().Sub(ts_begin))
}

func (i *Image) Close() error {
	return i.buffer.Close()
}

func (i *Image) Read(p []byte) (int, error) {
	if i.buffer == nil {
		return 0, fmt.Errorf("Bild muss erst geladen werden")
	}

	return i.buffer.Read(p)
}
