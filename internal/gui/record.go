package gui

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
)

type recorder struct {
	frames    []*image.Paletted
	delayCs   int
	scale     int
	maxFrames int
	done      bool
}

func newRecorder(maxFrames, fps, scale int) *recorder {
	return &recorder{
		delayCs:   max(100/max(fps, 1), 1),
		scale:     max(scale, 1),
		maxFrames: max(maxFrames, 1),
	}
}

func (r *recorder) add(img image.Image) {
	if r.done {
		return
	}
	small := downscale(img, r.scale)
	p := image.NewPaletted(small.Bounds(), palette.Plan9)
	draw.FloydSteinberg.Draw(p, small.Bounds(), small, image.Point{})
	r.frames = append(r.frames, p)
	if len(r.frames) >= r.maxFrames {
		r.done = true
	}
}

func (r *recorder) save(path string) error {
	if len(r.frames) == 0 {
		return fmt.Errorf("recorder: no frames captured")
	}
	delays := make([]int, len(r.frames))
	for i := range delays {
		delays[i] = r.delayCs
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("recorder: create %q: %w", path, err)
	}
	if err := gif.EncodeAll(f, &gif.GIF{Image: r.frames, Delay: delays}); err != nil {
		_ = f.Close()
		return fmt.Errorf("recorder: encode %q: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("recorder: close %q: %w", path, err)
	}
	return nil
}

func downscale(img image.Image, factor int) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx()/factor, b.Dy()/factor
	out := image.NewRGBA(image.Rect(0, 0, max(w, 1), max(h, 1)))
	for y := range out.Bounds().Dy() {
		for x := range out.Bounds().Dx() {
			out.Set(x, y, img.At(b.Min.X+x*factor, b.Min.Y+y*factor))
		}
	}
	return out
}
