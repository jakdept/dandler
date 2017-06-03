package handler

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"github.com/traherom/memstream"
)

// SplitHandler allows the routing of one handler at /, and another at all locations below /.
func SplitHandler(root, more http.Handler) http.Handler {
	return splitHandler{bare: root, more: more}
}

type splitHandler struct {
	bare http.Handler
	more http.Handler
}

func (p splitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if path.Clean(r.URL.Path) == "/" {
		p.bare.ServeHTTP(w, r)
	} else {
		p.more.ServeHTTP(w, r)
	}
}

// ContentTypeHandler serves a given file back to the requester, and determines content type by algorithm only.
// It does not use the file's extension to determine the content type.
func ContentTypeHandler(logger *log.Logger, basePath string) http.Handler {
	return contentTypeHandler{basePath: basePath, l: logger}
}

type contentTypeHandler struct {
	basePath string
	l        *log.Logger
}

// contentTypeHandler.ServeHTTP satasifies the Handler interface.
func (c contentTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(c.basePath, r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
		c.l.Printf("404 - could not open file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - could not stat file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	chunk := make([]byte, 512)

	_, err = f.Read(chunk)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - could not read from file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - could not seek within file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(chunk))
	http.ServeContent(w, r, r.URL.Path, stat.ModTime(), f)

	return
}

func ThumbnailHandler(logger *log.Logger, targetWidth, targetHeight int,
	rawImageDirectory, thumbnailDirectory, thumbnailExtension string) http.Handler {
	return thumbnailHandler{
		x:        targetWidth,
		y:        targetHeight,
		raw:      rawImageDirectory,
		thumbs:   thumbnailDirectory,
		thumbExt: thumbnailExtension,
		l:        logger,
	}
}

type thumbnailHandler struct {
	x        int
	y        int
	raw      string
	thumbs   string
	thumbExt string
	l        *log.Logger
}

func (h thumbnailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(h.generateThumbPath(h.trimThumbExt(r.URL.Path)))
	if err == nil {
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
			h.l.Printf("500 - could not stat file: %s - %s", filepath.Join(h.thumbs, r.URL.Path), err)
			return
		}

		w.Header().Set("Content-Type", "image/"+h.thumbExt)
		http.ServeContent(w, r, r.URL.Path, stat.ModTime(), f)
		return
	}

	if !os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		h.l.Printf("500 - error opening file: %s - %s", filepath.Join(h.thumbs, r.URL.Path), err)
		return
	}

	var img image.Image
	img, err = h.loadThumbnail(h.trimThumbExt(r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusNotFound)
		h.l.Printf("500 - error opening file: %s - %s", filepath.Join(h.thumbs, r.URL.Path), err)
		return
	}

	buf := memstream.NewCapacity(1000000)
	// rewrite to just generate an Encoder, and use that later maybe instead?
	w.Header().Set("Content-Type", "image/"+h.thumbExt)
	switch h.thumbExt {
	case "jpg":
		jpeg.Encode(buf, img, nil)
	case "jpeg":
		jpeg.Encode(buf, img, nil)
	case "png":
		png.Encode(buf, img)
	default:
		http.Error(w, fmt.Sprintf("could not respond with file; %s", r.URL.Path), http.StatusInternalServerError)
		h.l.Printf("500 - error pushing thumbnail: %s - %s", filepath.Join(h.thumbs, r.URL.Path), err)
		return
	}

	buf.Rewind()
	http.ServeContent(w, r, r.URL.Path, time.Now(), buf)
}

func (h thumbnailHandler) loadThumbnail(imageName string) (image.Image, error) {
	img, format, err := h.openImage(h.generateThumbPath(imageName))
	if os.IsNotExist(err) || format != h.thumbExt {
		img, _, err = h.openImage(h.generateRawPath(imageName))
		if err != nil {
			return nil, fmt.Errorf("could not open image [%s]: %s", imageName, err)
		}
		img, err = h.generateThumbnail(img)
		if err != nil {
			return nil, fmt.Errorf("could not process [%s]: %s", imageName, err)
		}
		err = h.writeThumbnail(imageName, img)
		if err != nil {
			return nil, fmt.Errorf("could not cache [%s]: %s", imageName, err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("problem loading thumbnail [%s]: %s", imageName, err)
	}
	return img, nil
}

func (h thumbnailHandler) writeThumbnail(imageName string, thumbnailImage image.Image) error {
	out, err := os.Create(h.generateThumbPath(imageName))
	if err != nil {
		return err
	}
	defer out.Close()
	switch h.thumbExt {
	case "jpg":
		return jpeg.Encode(out, thumbnailImage, nil)
	case "jpeg":
		return jpeg.Encode(out, thumbnailImage, nil)
	case "png":
		return png.Encode(out, thumbnailImage)
	default:
		return fmt.Errorf("extension type [%s] not supported for thumbnails", h.thumbExt)
	}
}

func (h thumbnailHandler) generateThumbnail(rawImage image.Image) (image.Image, error) {
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

func (h thumbnailHandler) openImage(imageName string) (image.Image, string, error) {
	path := filepath.Clean(imageName)
	reader, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

func (h thumbnailHandler) generateThumbPath(imageName string) string {
	return path.Clean(fmt.Sprintf("%s/%s.%s", h.thumbs, imageName, h.thumbExt))
}

func (h thumbnailHandler) generateRawPath(imageName string) string {
	return path.Clean(fmt.Sprintf("%s/%s", h.raw, imageName))
}

func (h thumbnailHandler) trimThumbExt(in string) string {
	return path.Clean(strings.TrimSuffix(in, "."+h.thumbExt))
}
