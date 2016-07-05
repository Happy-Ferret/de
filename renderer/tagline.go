package renderer

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	"github.com/driusan/de/demodel"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// TaglineRenderer is like the default renderer which doesn't
// provide syntax highlighting, but uses a different background
// colour which doesn't change regardless of mode.
type TaglineRenderer struct {
	DefaultSizeCalcer
	DefaultImageMapper
}

func (r *TaglineRenderer) InvalidateCache() {
	r.DefaultSizeCalcer.InvalidateCache()
	r.DefaultImageMapper.InvalidateCache()
}

func (r TaglineRenderer) CanRender(*demodel.CharBuffer) bool {
	return false
}

func (r TaglineRenderer) RenderInto(dst draw.Image, buf *demodel.CharBuffer, viewport image.Rectangle) error {
	dstSize := r.Bounds(buf)
	if wSize := viewport.Size().X; dstSize.Max.X < wSize {
		dstSize.Max.X = wSize
	}

	draw.Draw(dst, dst.Bounds(), &image.Uniform{TaglineBackground}, image.ZP, draw.Src)
	draw.Draw(dst,
		image.Rectangle{
			Min: image.Point{0, dstSize.Max.Y - 1},
			Max: image.Point{dstSize.Max.X, dstSize.Max.Y},
		},
		&image.Uniform{color.Black},
		image.ZP,
		draw.Src,
	)

	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{color.Black},
		Face: MonoFontFace,
		Dot:  fixed.P(0, MonoFontAscent.Floor()),
	}
	runes := bytes.Runes(buf.Buffer)

	for i, r := range runes {
		runeRectangle := image.Rectangle{}
		runeRectangle.Min.X = writer.Dot.X.Ceil()
		runeRectangle.Min.Y = writer.Dot.Y.Ceil() - MonoFontAscent.Floor()
		switch r {
		case '\t':
			runeRectangle.Max.X = runeRectangle.Min.X + 8*MonoFontGlyphWidth.Ceil()
		case '\n':
			runeRectangle.Max.X = dstSize.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + MonoFontGlyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + MonoFontHeight.Ceil()

		if buf.Dot.Start != buf.Dot.End && uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
			writer.Src = &image.Uniform{color.Black}
			draw.Draw(
				dst,
				runeRectangle,
				&image.Uniform{TextHighlight},
				image.ZP,
				draw.Over,
			)
		} else {

			if buf.Dot.Start == buf.Dot.End && buf.Dot.Start == uint(i) {
				// draw a cursor at the current location.
				draw.Draw(dst,
					image.Rectangle{
						image.Point{runeRectangle.Min.X, runeRectangle.Min.Y},
						image.Point{runeRectangle.Min.X + 1, runeRectangle.Max.Y},
					},
					&image.Uniform{color.Black},
					image.ZP,
					draw.Over,
				)
			}

			// it's not within dot, so use a black font on no background.
			writer.Src = &image.Uniform{color.Black}
		}
		switch r {
		case '\t':
			writer.Dot.X += MonoFontGlyphWidth * 8
			continue
		case '\n':
			writer.Dot.Y += MonoFontHeight
			writer.Dot.X = 0
			continue
		}
		writer.DrawString(string(r))
	}

	return nil
}
