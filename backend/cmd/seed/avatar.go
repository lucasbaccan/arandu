package main

import (
	"bytes"
	"encoding/base64"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"
)

// avatarDataURL gera um avatar tipo "identicon" (grade simétrica colorida,
// como GitHub gera pra usuários sem foto) determinístico a partir de um
// seed (o e-mail do participante), e o devolve como data URL — mesmo
// formato (data:image/...) que o upload real de foto grava em
// participants.photo (ver AvatarCropper.svelte / handleUpdateParticipantPhoto).
// Serve como asset fake pros participantes "com foto" do seed, sem precisar
// de nenhum arquivo de imagem externo.
func avatarDataURL(seed string) string {
	const gridSize = 5
	const cellPx = 32
	const uniqueCols = (gridSize + 1) / 2 // colunas únicas antes do espelhamento horizontal

	h := fnv.New64a()
	_, _ = h.Write([]byte(seed))
	sum := h.Sum64()

	fg := hslToRGBA(float64(sum%360), 0.55, 0.55)
	bg := color.RGBA{R: 0xf1, G: 0xf1, B: 0xf1, A: 0xff}

	size := gridSize * cellPx
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
		}
	}

	bits := sum
	for row := 0; row < gridSize; row++ {
		for col := 0; col < uniqueCols; col++ {
			on := bits&1 == 1
			bits >>= 1
			if !on {
				continue
			}
			fillCell(img, row, col, cellPx, fg)
			mirrorCol := gridSize - 1 - col
			if mirrorCol != col {
				fillCell(img, row, mirrorCol, cellPx, fg)
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func fillCell(img *image.RGBA, row, col, cellPx int, c color.RGBA) {
	x0, y0 := col*cellPx, row*cellPx
	for y := y0; y < y0+cellPx; y++ {
		for x := x0; x < x0+cellPx; x++ {
			img.Set(x, y, c)
		}
	}
}

// hslToRGBA converte HSL (h em graus 0-360, s/l em 0-1) pra RGBA opaco.
func hslToRGBA(h, s, l float64) color.RGBA {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}
