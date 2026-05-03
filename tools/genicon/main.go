// Command genicon builds build/windows/icon.ico from the application logo.
// Run with: `go run ./tools/genicon`. Pixel-art-friendly nearest-neighbour
// scaling is used so the icon stays crisp at small sizes.
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"log"
	"os"
)

const (
	logoPath = "frontend/src/assets/images/logo.png"
	icoPath  = "build/windows/icon.ico"
)

// Standard Windows icon resolutions. 256 sits in the .ico header as 0
// (the format uses a single byte for width/height and 256 doesn't fit).
var sizes = []int{256, 128, 64, 48, 32, 16}

func main() {
	src, err := loadPNG(logoPath)
	if err != nil {
		log.Fatalf("load logo: %v", err)
	}

	encoded := make([][]byte, len(sizes))
	for i, sz := range sizes {
		scaled := scale(src, sz, sz)
		var buf bytes.Buffer
		if err := png.Encode(&buf, scaled); err != nil {
			log.Fatalf("encode %dx%d: %v", sz, sz, err)
		}
		encoded[i] = buf.Bytes()
	}

	out, err := os.Create(icoPath)
	if err != nil {
		log.Fatalf("create %s: %v", icoPath, err)
	}
	defer out.Close()

	// ICONDIR header (6 bytes).
	bw := binary.Write
	if err := bw(out, binary.LittleEndian, uint16(0)); err != nil {
		log.Fatal(err)
	}
	if err := bw(out, binary.LittleEndian, uint16(1)); err != nil { // 1 = icon
		log.Fatal(err)
	}
	if err := bw(out, binary.LittleEndian, uint16(len(sizes))); err != nil {
		log.Fatal(err)
	}

	offset := 6 + 16*len(sizes)
	for i, sz := range sizes {
		dim := uint8(sz)
		if sz == 256 {
			dim = 0
		}
		entry := []any{
			dim,                       // width
			dim,                       // height
			uint8(0),                  // colour palette size (0 = no palette)
			uint8(0),                  // reserved
			uint16(1),                 // colour planes
			uint16(32),                // bits per pixel
			uint32(len(encoded[i])),   // image data size
			uint32(offset),            // image data offset
		}
		for _, v := range entry {
			if err := binary.Write(out, binary.LittleEndian, v); err != nil {
				log.Fatal(err)
			}
		}
		offset += len(encoded[i])
	}

	for _, blob := range encoded {
		if _, err := out.Write(blob); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("wrote %s with sizes %v", icoPath, sizes)
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

// scale resamples src to wxh using nearest-neighbour. Suitable for pixel
// art where bilinear/bicubic would blur the hard edges.
func scale(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	for y := 0; y < h; y++ {
		sy := y * sh / h
		for x := 0; x < w; x++ {
			sx := x * sw / w
			dst.Set(x, y, src.At(sb.Min.X+sx, sb.Min.Y+sy))
		}
	}
	return dst
}
