// Package ebiten provides the windowed Renderer built on the Ebiten game
// library. The implementation is compiled only with the "ebiten" build tag,
// because it requires a display and OpenGL; this file keeps the package
// buildable (and the rest of the module testable) without that tag.
package ebiten
