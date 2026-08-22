package handlers

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"strings"
	"testing"
)

func TestAgencyQRRejectsUnsafeInputs(t *testing.T) {
	for _, input := range []string{"data:image/svg+xml;base64,PHN2Zy8+", "data:image/png;base64,notbase64", "https://example.test/image.png", "data:image/png;base64," + strings.Repeat("A", 700001)} {
		if _, err := decodeAgencyQR(input); err == nil {
			t.Fatal("invalid image was accepted")
		}
	}
	var raw bytes.Buffer
	_ = png.Encode(&raw, image.NewGray(image.Rect(0, 0, 2049, 1)))
	if _, err := decodeAgencyQR("data:image/png;base64," + base64.StdEncoding.EncodeToString(raw.Bytes())); err == nil {
		t.Fatal("oversized dimensions accepted")
	}
	if data, err := decodeAgencyQR(""); err != nil || data != nil {
		t.Fatal("empty QR should be optional")
	}
}
