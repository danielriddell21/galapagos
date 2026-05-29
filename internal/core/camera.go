package core

// Camera maps world coordinates to screen coordinates with pan and zoom, and
// can follow a moving target. It is a pure value type so the rendering backend
// only reads it; all camera math is testable without a display.
type Camera struct {
	// X, Y is the world point shown at the center of the viewport.
	X, Y float64
	// Zoom is the scale factor; values above 1 zoom in.
	Zoom float64
	// ViewW, ViewH are the viewport dimensions in screen pixels.
	ViewW, ViewH float64
}

// NewCamera returns a camera centered at the origin with unit zoom for a
// viewport of the given pixel size.
func NewCamera(viewW, viewH float64) *Camera {
	return &Camera{Zoom: 1, ViewW: viewW, ViewH: viewH}
}

// WorldToScreen converts a world point to screen pixel coordinates.
func (c *Camera) WorldToScreen(wx, wy float64) (sx, sy float64) {
	sx = (wx-c.X)*c.Zoom + c.ViewW/2
	sy = (wy-c.Y)*c.Zoom + c.ViewH/2
	return sx, sy
}

// ScreenToWorld converts screen pixel coordinates to a world point. It is the
// inverse of [Camera.WorldToScreen].
func (c *Camera) ScreenToWorld(sx, sy float64) (wx, wy float64) {
	wx = (sx-c.ViewW/2)/c.Zoom + c.X
	wy = (sy-c.ViewH/2)/c.Zoom + c.Y
	return wx, wy
}

// Follow centers the camera on the given world point.
func (c *Camera) Follow(wx, wy float64) {
	c.X, c.Y = wx, wy
}

// FitBounds positions and zooms the camera so the axis-aligned world rectangle
// [minX,maxX] x [minY,maxY] fills the viewport, leaving a fractional margin.
func (c *Camera) FitBounds(minX, minY, maxX, maxY, margin float64) {
	c.X = (minX + maxX) / 2
	c.Y = (minY + maxY) / 2
	w := (maxX - minX) * (1 + margin)
	h := (maxY - minY) * (1 + margin)
	if w <= 0 || h <= 0 {
		c.Zoom = 1
		return
	}
	c.Zoom = min(c.ViewW/w, c.ViewH/h)
}
