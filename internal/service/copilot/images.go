// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif" // registers the GIF decoder for image.Decode
	"image/jpeg"
	_ "image/png" // registers the PNG decoder for image.Decode
	"net/http"
	"path"
	"strings"
	"sync"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // registers the WebP decoder for image.Decode
)

// What a vision request can carry, per the providers' published limits:
// Anthropic and OpenAI both accept exactly these four formats, Anthropic caps a
// single image at 5 MB and a whole request at 32 MB, and both scale anything
// past roughly 1.5k pixels on the long edge down before the model sees it.
const (
	// maxImageEdge is Anthropic's recommended long edge; larger images are
	// downscaled server-side anyway, so sending more only costs upload.
	maxImageEdge = 1568

	// maxEncodedImageBytes keeps one image under the 5 MB per-image limit
	// once base64 has grown it by a third.
	maxEncodedImageBytes = 5_000_000

	// runImageBudgetBytes bounds what all images in one run add to the
	// request, leaving the 32 MB ceiling for everything else.
	runImageBudgetBytes = 16_000_000

	jpegQuality = 85
)

var visionMediaTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// sniffMediaType names an image by its bytes, not by the stored Content-Type.
// The stored type can be empty or wrong, and a mismatch between declared and
// actual format is a 400 from the provider, not a degraded answer.
func sniffMediaType(data []byte) (string, bool) {
	mt := http.DetectContentType(data)
	return mt, visionMediaTypes[mt]
}

// mediaTypeOfURL guesses an external image's type from its declared type or
// extension. Only a recognised vision format is sent; the provider fetches the
// URL itself and fails the whole request on anything else.
func mediaTypeOfURL(declared, url string) (string, bool) {
	if mt := strings.ToLower(strings.TrimSpace(declared)); visionMediaTypes[mt] {
		return mt, true
	}
	switch strings.ToLower(path.Ext(strings.SplitN(url, "?", 2)[0])) {
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".png":
		return "image/png", true
	case ".gif":
		return "image/gif", true
	case ".webp":
		return "image/webp", true
	}
	return "", false
}

// fitForVision returns an image the providers will take, downscaling and
// re-encoding when it is larger than they would use. It gives up — the image
// is simply not attached — when the format is unsupported, or when the image
// cannot be decoded and is already too large to send as it is.
func fitForVision(data []byte) (mediaType string, out []byte, ok bool) {
	mediaType, ok = sniffMediaType(data)
	if !ok {
		return "", nil, false
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	fits := len(data)*4/3 <= maxEncodedImageBytes
	if err != nil {
		return mediaType, data, fits
	}
	if max(cfg.Width, cfg.Height) <= maxImageEdge && fits {
		return mediaType, data, true
	}

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return mediaType, data, fits
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, downscale(src, maxImageEdge), &jpeg.Options{Quality: jpegQuality}); err != nil {
		return mediaType, data, fits
	}
	if buf.Len()*4/3 > maxEncodedImageBytes {
		return "", nil, false
	}
	return "image/jpeg", buf.Bytes(), true
}

// downscale shrinks src so its long edge is at most edge, with Catmull-Rom
// resampling. Transparency is flattened onto white first, since the result is
// encoded as JPEG and would otherwise turn transparent areas black.
func downscale(src image.Image, edge int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	if w >= h && w > edge {
		dw, dh = edge, max(h*edge/w, 1)
	} else if h > w && h > edge {
		dw, dh = max(w*edge/h, 1), edge
	}

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

// imageBudget is what images may still add to one run's requests. Searches in
// the same round run in parallel, hence the lock.
type imageBudget struct {
	mu   sync.Mutex
	left int
}

func newImageBudget() *imageBudget { return &imageBudget{left: runImageBudgetBytes} }

// take reserves n encoded bytes, or reports that they no longer fit.
func (b *imageBudget) take(n int) bool {
	if b == nil {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if n > b.left {
		return false
	}
	b.left -= n
	return true
}
