package render

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func TextPNG(text string) ([]byte, error) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", "    ")
		lines[i] = line
	}

	const scale = 2
	fontData, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    13 * scale,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	defer face.Close()

	padding := 12 * scale
	maxWidth := 0
	for _, line := range lines {
		w := font.MeasureString(face, line).Ceil()
		if w > maxWidth {
			maxWidth = w
		}
	}
	lineHeight := face.Metrics().Height.Ceil()
	width := maxWidth + padding*2
	height := lineHeight*len(lines) + padding*2
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{R: 18, G: 18, B: 22, A: 255}
	fg := color.RGBA{R: 230, G: 232, B: 235, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fg),
		Face: face,
	}
	startY := padding + face.Metrics().Ascent.Ceil()
	for _, line := range lines {
		d.Dot = fixed.P(padding, startY)
		d.DrawString(line)
		startY += lineHeight
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
