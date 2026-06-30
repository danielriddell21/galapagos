package core

type Camera struct {
	X, Y float64

	Zoom float64

	ViewW, ViewH float64
}

func NewCamera(viewW, viewH float64) *Camera {
	return &Camera{Zoom: 1, ViewW: viewW, ViewH: viewH}
}

func (c *Camera) WorldToScreen(wx, wy float64) (sx, sy float64) {
	sx = (wx-c.X)*c.Zoom + c.ViewW/2
	sy = (wy-c.Y)*c.Zoom + c.ViewH/2
	return sx, sy
}

func (c *Camera) ScreenToWorld(sx, sy float64) (wx, wy float64) {
	wx = (sx-c.ViewW/2)/c.Zoom + c.X
	wy = (sy-c.ViewH/2)/c.Zoom + c.Y
	return wx, wy
}

func (c *Camera) Follow(wx, wy float64) {
	c.X, c.Y = wx, wy
}

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
