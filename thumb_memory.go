package dandler

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	_ "image/gif"  // imported to allow gif decoding natively
	_ "image/jpeg" // imported to allow jpeg decoding
	_ "image/png"  // imported to allow png decoding

	"image/jpeg"
	"image/png"

	"github.com/golang/groupcache"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

// ThumbCache returns a handler that serves thumbnails from GroupCache.
// Thumbnails are generated when needed by GroupCache.
func ThumbCache(logger *log.Logger, targetWidth, targetHeight int, cacheSize int64,
	rawImageDirectory, cacheName, thumbnailExtension string) http.Handler {
	this := thumbCache{
		x:        targetWidth,
		y:        targetHeight,
		raw:      rawImageDirectory,
		thumbExt: thumbnailExtension,
		l:        logger,
	}
	this.cache = groupcache.NewGroup(cacheName, cacheSize, this)
	return this
}

const Megabyte int = 2 << 20

type thumbCache struct {
	x        int
	y        int
	raw      string
	thumbExt string
	l        *log.Logger
	cache    *groupcache.Group
}

func (h thumbCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/"+h.thumbExt)
	data := new([]byte)
	err := h.cache.Get(nil, r.URL.Path, groupcache.AllocatingByteSliceSink(data))
	if err != nil {
		h.l.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(*data))
}

func (h thumbCache) Get(ctx groupcache.Context, key string,
	dest groupcache.Sink) error {
	value, err := h.generateThumbnail(key)
	if err != nil {
		return err
	}
	dest.SetBytes(value)
	return nil
}

func (h thumbCache) generateThumbnail(imageName string) ([]byte, error) {
	rawImage, err := h.openImage(h.generateRawPath(imageName))
	if err != nil {
		log.Println(err)
		return []byte{}, fmt.Errorf("could not open image [%s]: %s", imageName, err)
	}
	thumbImg, err := h.resizeImage(rawImage)
	if err != nil {
		log.Println(err)
		return []byte{}, fmt.Errorf("cound not resize image [%s]: %v", imageName, err)
	}
	data, err := h.encodeImage(thumbImg)
	return data, err
}

func (h thumbCache) openImage(imageName string) (image.Image, error) {
	path := filepath.Clean(imageName)
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (h thumbCache) resizeImage(rawImage image.Image) (image.Image, error) {
	shrunk := resize.Resize(0, uint(h.y), rawImage, resize.MitchellNetravali)
	thumbnail, err := cutter.Crop(shrunk, cutter.Config{
		Height:  h.y,
		Width:   h.x,
		Options: cutter.Copy,
		Mode:    cutter.Centered,
	})
	if err != nil {
		return nil, err
	}
	return thumbnail, nil
}

func (h thumbCache) encodeImage(img image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	switch h.thumbExt {
	case "jpg":
		jpeg.Encode(buf, img, nil)
	case "jpeg":
		jpeg.Encode(buf, img, nil)
	case "png":
		png.Encode(buf, img)
	default:
		return []byte{}, fmt.Errorf("extension [%s] not supported", h.thumbExt)
	}
	return buf.Bytes(), nil
}

func (h thumbCache) generateRawPath(imageName string) string {
	return path.Clean(fmt.Sprintf("%s/%s", h.raw, imageName))
}
