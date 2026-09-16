package handlers

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"testing"
)

func TestAttractionPhotoFiveMiBBoundary(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{2 << 20, 5 << 20, (5 << 20) + 1} {
		// PNG readers allow trailing bytes; the upload limit must count all bytes.
		raw := make([]byte, size)
		copy(raw, encoded.Bytes())
		_, err := decodeAttractionImage("data:image/png;base64," + base64.StdEncoding.EncodeToString(raw))
		if (err != nil) != (size > 5<<20) {
			t.Fatalf("size %d: error %v", size, err)
		}
	}
}
