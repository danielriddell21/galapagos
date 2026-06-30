package gui

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
)

func solidFrame(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestRecorderSavesGIF(t *testing.T) {
	r := newRecorder(3, 25, 2)
	colors := []color.RGBA{{200, 40, 40, 255}, {40, 170, 70, 255}, {40, 90, 200, 255}}
	for i, c := range colors {
		r.add(solidFrame(64, 48, c))
		if i < len(colors)-1 && r.done {
			t.Fatalf("recorder marked done early at frame %d", i)
		}
	}
	if !r.done {
		t.Fatal("recorder should be done after maxFrames")
	}
	r.add(solidFrame(64, 48, color.RGBA{0, 0, 0, 255})) // ignored once done
	if len(r.frames) != 3 {
		t.Fatalf("captured %d frames, want 3", len(r.frames))
	}

	path := filepath.Join(t.TempDir(), "demo.gif")
	if err := r.save(path); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	g, err := gif.DecodeAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Image) != 3 {
		t.Fatalf("decoded %d frames, want 3", len(g.Image))
	}
	// Downscaled by 2: 64x48 -> 32x24.
	if b := g.Image[0].Bounds(); b.Dx() != 32 || b.Dy() != 24 {
		t.Fatalf("frame size = %v, want 32x24", b.Size())
	}
	if g.Delay[0] == 0 {
		t.Fatal("frame delay should be non-zero")
	}
}

func TestRecorderSaveNoFrames(t *testing.T) {
	if err := newRecorder(5, 25, 1).save(filepath.Join(t.TempDir(), "x.gif")); err == nil {
		t.Fatal("saving with no frames should error")
	}
}
