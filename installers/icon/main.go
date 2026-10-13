// Command icon xera as iconas da aplicación a partir do mesmo deseño que
// frontend/public/favicon.svg: installers/windows/app.ico e
// installers/macos/AppIcon.icns. Execútase desde a raíz do repositorio:
//
//	go run ./installers/icon
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// Deseño do favicon nun lenzo de 64×64: cadrado redondeado verde cunha liña
// de base e outra de evolución, as dúas brancas e con extremos redondos.
const (
	canvas      = 64.0
	cornerR     = 14.0
	strokeHalfW = 2.5
)

var (
	green = color.NRGBA{R: 0x0c, G: 0xa6, B: 0x78, A: 0xff}
	white = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	segments = [][4]float64{
		{14, 46, 50, 46}, // liña de base
		{18, 39, 28, 29}, // evolución
		{28, 29, 36, 36},
		{36, 36, 48, 19},
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func run() error {
	// Windows: a icona ocupa todo o cadrado, coma o favicon.
	icoSizes := []int{16, 24, 32, 48, 64, 128, 256}
	icoImages := make([][]byte, 0, len(icoSizes))
	for _, size := range icoSizes {
		data, err := encodePNG(render(size, 0))
		if err != nil {
			return err
		}
		icoImages = append(icoImages, data)
	}
	if err := os.WriteFile("installers/windows/app.ico", buildICO(icoImages, icoSizes), 0o644); err != nil {
		return err
	}

	// macOS: as iconas levan unha marxe arredor (824 de 1024 px).
	const macPadding = (1 - 824.0/1024.0) / 2
	type icnsEntry struct {
		kind string
		size int
	}
	var icns []icnsElement
	for _, e := range []icnsEntry{
		{"icp4", 16}, {"icp5", 32}, {"ic11", 32}, {"ic12", 64}, {"ic07", 128},
		{"ic13", 256}, {"ic08", 256}, {"ic14", 512}, {"ic09", 512}, {"ic10", 1024},
	} {
		data, err := encodePNG(render(e.size, macPadding))
		if err != nil {
			return err
		}
		icns = append(icns, icnsElement{kind: e.kind, data: data})
	}
	if err := os.WriteFile("installers/macos/AppIcon.icns", buildICNS(icns), 0o644); err != nil {
		return err
	}
	fmt.Println("Xeradas installers/windows/app.ico e installers/macos/AppIcon.icns")
	return nil
}

// render debuxa a icona de size×size píxeles con supermostraxe (antialiasing).
// padding é a fracción do lado que queda baleira en cada bordo.
func render(size int, padding float64) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	n := 8
	if size > 128 {
		n = 4
	}
	content := float64(size) * (1 - 2*padding)
	offset := float64(size) * padding
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var r, g, b, a float64
			for sy := 0; sy < n; sy++ {
				for sx := 0; sx < n; sx++ {
					x := (float64(px) + (float64(sx)+0.5)/float64(n) - offset) / content * canvas
					y := (float64(py) + (float64(sy)+0.5)/float64(n) - offset) / content * canvas
					if !insideRoundedSquare(x, y) {
						continue
					}
					c := green
					if onStroke(x, y) {
						c = white
					}
					r += float64(c.R)
					g += float64(c.G)
					b += float64(c.B)
					a++
				}
			}
			if a == 0 {
				continue
			}
			img.SetNRGBA(px, py, color.NRGBA{
				R: uint8(math.Round(r / a)),
				G: uint8(math.Round(g / a)),
				B: uint8(math.Round(b / a)),
				A: uint8(math.Round(255 * a / float64(n*n))),
			})
		}
	}
	return img
}

func insideRoundedSquare(x, y float64) bool {
	if x < 0 || y < 0 || x > canvas || y > canvas {
		return false
	}
	cx := math.Max(cornerR, math.Min(x, canvas-cornerR))
	cy := math.Max(cornerR, math.Min(y, canvas-cornerR))
	return math.Hypot(x-cx, y-cy) <= cornerR
}

func onStroke(x, y float64) bool {
	for _, s := range segments {
		if distToSegment(x, y, s[0], s[1], s[2], s[3]) <= strokeHalfW {
			return true
		}
	}
	return false
}

func distToSegment(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildICO crea un .ico con imaxes PNG (formato admitido desde Windows Vista).
func buildICO(images [][]byte, sizes []int) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, [3]uint16{0, 1, uint16(len(images))})
	offset := 6 + 16*len(images)
	for i, data := range images {
		dim := uint8(sizes[i])
		if sizes[i] >= 256 {
			dim = 0 // 0 significa 256 no formato ICO
		}
		buf.Write([]byte{dim, dim, 0, 0})
		_ = binary.Write(&buf, binary.LittleEndian, [2]uint16{1, 32})
		_ = binary.Write(&buf, binary.LittleEndian, [2]uint32{uint32(len(data)), uint32(offset)})
		offset += len(data)
	}
	for _, data := range images {
		buf.Write(data)
	}
	return buf.Bytes()
}

type icnsElement struct {
	kind string
	data []byte
}

// buildICNS crea un .icns con imaxes PNG (formato de macOS 10.7 en diante).
func buildICNS(elements []icnsElement) []byte {
	var body bytes.Buffer
	for _, e := range elements {
		body.WriteString(e.kind)
		_ = binary.Write(&body, binary.BigEndian, uint32(8+len(e.data)))
		body.Write(e.data)
	}
	var buf bytes.Buffer
	buf.WriteString("icns")
	_ = binary.Write(&buf, binary.BigEndian, uint32(8+body.Len()))
	buf.Write(body.Bytes())
	return buf.Bytes()
}
