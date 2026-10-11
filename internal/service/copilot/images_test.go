// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

func encodePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: 0x80, A: 0xff})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFitForVision_SmallImagePassesThroughWithSniffedType(t *testing.T) {
	data := encodePNG(t, 64, 32)
	mt, out, ok := fitForVision(data)
	if !ok || mt != "image/png" || !bytes.Equal(out, data) {
		t.Fatalf("got %q ok=%v changed=%v; a small PNG should go as-is, typed by its bytes", mt, ok, !bytes.Equal(out, data))
	}
}

func TestFitForVision_LargeImageIsDownscaled(t *testing.T) {
	mt, out, ok := fitForVision(encodePNG(t, 3200, 1600))
	if !ok || mt != "image/jpeg" {
		t.Fatalf("got %q ok=%v, want a re-encoded JPEG", mt, ok)
	}
	img, err := jpeg.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != maxImageEdge || b.Dy() != maxImageEdge/2 {
		t.Fatalf("size = %dx%d, want %dx%d", b.Dx(), b.Dy(), maxImageEdge, maxImageEdge/2)
	}
}

// A WebP larger than the providers use is decoded and downscaled like any
// other format, rather than sent at full size or dropped.
func TestFitForVision_LargeWebPIsDownscaled(t *testing.T) {
	data, err := os.ReadFile("testdata/wide-2400x1200.webp")
	if err != nil {
		t.Fatal(err)
	}
	mt, out, ok := fitForVision(data)
	if !ok || mt != "image/jpeg" {
		t.Fatalf("got %q ok=%v, want a downscaled JPEG", mt, ok)
	}
	img, err := jpeg.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != maxImageEdge || b.Dy() != maxImageEdge/2 {
		t.Fatalf("size = %dx%d, want %dx%d", b.Dx(), b.Dy(), maxImageEdge, maxImageEdge/2)
	}
}

func TestFitForVision_UnsupportedFormatIsDropped(t *testing.T) {
	// An AVIF upload is allowed by the site but not by any vision API; one
	// such image must not fail the whole request.
	avif := append([]byte{0, 0, 0, 0x1c}, []byte("ftypavif")...)
	if _, _, ok := fitForVision(append(avif, make([]byte, 64)...)); ok {
		t.Fatalf("AVIF must not be sent")
	}
}

func TestMediaTypeOfURL(t *testing.T) {
	for url, want := range map[string]string{
		"https://x/a.JPG?w=1": "image/jpeg",
		"https://x/b.webp":    "image/webp",
		"https://x/c.avif":    "",
		"https://x/d":         "",
	} {
		got, ok := mediaTypeOfURL("", url)
		if got != want || ok != (want != "") {
			t.Fatalf("%s: got %q ok=%v, want %q", url, got, ok, want)
		}
	}
	if got, ok := mediaTypeOfURL("image/png", "https://x/d"); !ok || got != "image/png" {
		t.Fatalf("a declared vision type should be trusted for external URLs")
	}
}

func TestImageBudget(t *testing.T) {
	b := &imageBudget{left: 10}
	if !b.take(6) || b.take(6) || !b.take(4) {
		t.Fatalf("budget should admit 6, refuse the next 6, then admit 4")
	}
	var unlimited *imageBudget
	if !unlimited.take(1 << 30) {
		t.Fatalf("a nil budget places no limit")
	}
}
