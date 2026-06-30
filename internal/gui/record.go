package gui

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
)

// recorder accumulates downscaled, palette-quantized frames and writes them as an
// animated GIF. It is independent of the graphics backend (it takes plain
// image.Image frames), so it is unit-tested without a display.
type recorder struct {
	frames    []*image.Paletted
	delayCs   int // per-frame delay in hundredths of a second
	scale     int // integer downscale factor (1 = full size)
	maxFrames int
	done      bool
}

// newRecorder returns a recorder capturing up to maxFrames at fps, downscaled by
// scale.
func newRecorder(maxFrames, fps, scale int) *recorder {
	return &recorder{
		delayCs:   max(100/max(fps, 1), 1),
		scale:     max(scale, 1),
		maxFrames: max(maxFrames, 1),
	}
}

// add downscales and quantizes a frame, appending it until maxFrames is reached.
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

// save writes the accumulated frames as an animated GIF.
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

// downscale shrinks img by an integer factor using nearest-neighbour sampling. A
// factor of 1 returns an RGBA copy unchanged.
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
