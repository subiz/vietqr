package vietqr

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"path"
	"strings"
	"sync"

	"github.com/subiz/vietqr/internal/qrcode"
)

// PNGWidth and PNGHeight are the fixed dimensions of generated VietQR images.
const (
	PNGWidth  = 850
	PNGHeight = 1100
)

var (
	canvasWhite = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	vietQRBlue  = color.RGBA{R: 25, G: 59, B: 126, A: 255}
)

//go:embed image/vietqr.png image/napas247.png image/banks/*.png image/banks/index.json
var pngAssets embed.FS

type logoManifest struct {
	Logos []logoManifestEntry `json:"logos"`
}

type logoManifestEntry struct {
	Code      string `json:"code"`
	BIN       string `json:"bin"`
	ShortName string `json:"shortName"`
	File      string `json:"file"`
}

type logoLookup struct {
	byBIN       map[string]string
	byCode      map[string]string
	byShortName map[string]string
}

var (
	logoLookupOnce sync.Once
	logos          logoLookup
	logoLookupErr  error
)

// GeneratePNG generates a complete 850x1100 VietQR image for a typical VND
// account transfer. The returned PNG contains the VietQR logo, a framed QR
// code, and the NAPAS 247 and beneficiary bank logos.
func GeneratePNG(amount float64, bankBIN, accountNumber, note string) ([]byte, error) {
	return RenderPNG(Generate(amount, bankBIN, accountNumber, note), bankBIN)
}

// RenderPNG renders an existing VietQR payload as an 850x1100 PNG. bankBIN is
// used only to select the beneficiary bank logo embedded in this package.
func RenderPNG(payload, bankBIN string) ([]byte, error) {
	if strings.TrimSpace(payload) == "" {
		return nil, fmt.Errorf("render VietQR PNG: empty payload")
	}

	bankLogo, err := loadBankLogo(bankBIN)
	if err != nil {
		return nil, fmt.Errorf("render VietQR PNG: %w", err)
	}
	vietQRLogo, err := decodePNGAsset("image/vietqr.png")
	if err != nil {
		return nil, fmt.Errorf("render VietQR PNG: %w", err)
	}
	napasLogo, err := decodePNGAsset("image/napas247.png")
	if err != nil {
		return nil, fmt.Errorf("render VietQR PNG: %w", err)
	}

	code, err := qrcode.New(payload, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("render VietQR PNG: encode QR: %w", err)
	}
	code.DisableBorder = true

	canvas := image.NewRGBA(image.Rect(0, 0, PNGWidth, PNGHeight))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(canvasWhite), image.Point{}, draw.Src)

	drawFitted(canvas, vietQRLogo, image.Rect(175, 45, 675, 170))

	qrFrame := image.Rect(75, 190, 775, 890)
	drawBorder(canvas, qrFrame, 3, vietQRBlue)
	if err := drawQRBitmap(canvas, code.Bitmap(), qrFrame.Inset(20)); err != nil {
		return nil, fmt.Errorf("render VietQR PNG: %w", err)
	}

	drawFitted(canvas, napasLogo, image.Rect(70, 950, 390, 1035))
	draw.Draw(canvas, image.Rect(424, 935, 427, 1045), image.NewUniform(vietQRBlue), image.Point{}, draw.Src)
	drawFitted(canvas, bankLogo, image.Rect(465, 935, 785, 1045))

	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		return nil, fmt.Errorf("render VietQR PNG: encode PNG: %w", err)
	}
	return output.Bytes(), nil
}

func drawQRBitmap(dst draw.Image, bitmap [][]bool, bounds image.Rectangle) error {
	if len(bitmap) == 0 || len(bitmap[0]) == 0 {
		return fmt.Errorf("empty QR bitmap")
	}
	moduleCount := len(bitmap)
	for _, row := range bitmap {
		if len(row) != moduleCount {
			return fmt.Errorf("non-square QR bitmap")
		}
	}

	size := min(bounds.Dx(), bounds.Dy())
	if size < moduleCount {
		return fmt.Errorf("QR bitmap with %d modules does not fit", moduleCount)
	}
	startX := bounds.Min.X + (bounds.Dx()-size)/2
	startY := bounds.Min.Y + (bounds.Dy()-size)/2
	black := image.NewUniform(color.Black)

	for y, row := range bitmap {
		for x, dark := range row {
			if !dark {
				continue
			}
			module := image.Rect(
				startX+x*size/moduleCount,
				startY+y*size/moduleCount,
				startX+(x+1)*size/moduleCount,
				startY+(y+1)*size/moduleCount,
			)
			draw.Draw(dst, module, black, image.Point{}, draw.Src)
		}
	}
	return nil
}

func drawBorder(dst draw.Image, bounds image.Rectangle, width int, border color.Color) {
	uniform := image.NewUniform(border)
	draw.Draw(dst, image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+width), uniform, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(bounds.Min.X, bounds.Max.Y-width, bounds.Max.X, bounds.Max.Y), uniform, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+width, bounds.Max.Y), uniform, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(bounds.Max.X-width, bounds.Min.Y, bounds.Max.X, bounds.Max.Y), uniform, image.Point{}, draw.Src)
}

func drawFitted(dst draw.Image, source image.Image, bounds image.Rectangle) {
	contentBounds := visibleBounds(source)
	width, height := fittedSize(contentBounds.Dx(), contentBounds.Dy(), bounds.Dx(), bounds.Dy())
	if width == 0 || height == 0 {
		return
	}

	resized := resizeBilinear(source, contentBounds, width, height)
	target := image.Rect(
		bounds.Min.X+(bounds.Dx()-width)/2,
		bounds.Min.Y+(bounds.Dy()-height)/2,
		bounds.Min.X+(bounds.Dx()-width)/2+width,
		bounds.Min.Y+(bounds.Dy()-height)/2+height,
	)
	draw.Draw(dst, target, resized, image.Point{}, draw.Over)
}

func fittedSize(sourceWidth, sourceHeight, maxWidth, maxHeight int) (int, int) {
	if sourceWidth <= 0 || sourceHeight <= 0 || maxWidth <= 0 || maxHeight <= 0 {
		return 0, 0
	}
	scale := math.Min(float64(maxWidth)/float64(sourceWidth), float64(maxHeight)/float64(sourceHeight))
	return max(1, int(math.Round(float64(sourceWidth)*scale))), max(1, int(math.Round(float64(sourceHeight)*scale)))
}

func visibleBounds(source image.Image) image.Rectangle {
	bounds := source.Bounds()
	visible := image.Rectangle{}
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			red, green, blue, alpha := source.At(x, y).RGBA()
			if alpha < 0x0800 {
				continue
			}
			red = min(uint32(0xffff), uint32(red)*0xffff/uint32(alpha))
			green = min(uint32(0xffff), uint32(green)*0xffff/uint32(alpha))
			blue = min(uint32(0xffff), uint32(blue)*0xffff/uint32(alpha))
			if red > 0xf500 && green > 0xf500 && blue > 0xf500 {
				continue
			}

			pixel := image.Rect(x, y, x+1, y+1)
			if !found {
				visible = pixel
				found = true
			} else {
				visible = visible.Union(pixel)
			}
		}
	}
	if !found {
		return bounds
	}
	return visible
}

func resizeBilinear(source image.Image, sourceBounds image.Rectangle, width, height int) *image.RGBA {
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	sourceWidth := sourceBounds.Dx()
	sourceHeight := sourceBounds.Dy()

	for y := 0; y < height; y++ {
		sourceY := (float64(y)+0.5)*float64(sourceHeight)/float64(height) - 0.5
		y0 := int(math.Floor(sourceY))
		yWeight := sourceY - float64(y0)
		if y0 < 0 {
			y0 = 0
			yWeight = 0
		}
		y1 := min(y0+1, sourceHeight-1)

		for x := 0; x < width; x++ {
			sourceX := (float64(x)+0.5)*float64(sourceWidth)/float64(width) - 0.5
			x0 := int(math.Floor(sourceX))
			xWeight := sourceX - float64(x0)
			if x0 < 0 {
				x0 = 0
				xWeight = 0
			}
			x1 := min(x0+1, sourceWidth-1)

			red, green, blue, alpha := bilinearRGBA(
				source.At(sourceBounds.Min.X+x0, sourceBounds.Min.Y+y0),
				source.At(sourceBounds.Min.X+x1, sourceBounds.Min.Y+y0),
				source.At(sourceBounds.Min.X+x0, sourceBounds.Min.Y+y1),
				source.At(sourceBounds.Min.X+x1, sourceBounds.Min.Y+y1),
				xWeight,
				yWeight,
			)
			resized.SetRGBA(x, y, color.RGBA{R: red, G: green, B: blue, A: alpha})
		}
	}
	return resized
}

func bilinearRGBA(topLeft, topRight, bottomLeft, bottomRight color.Color, xWeight, yWeight float64) (byte, byte, byte, byte) {
	tlR, tlG, tlB, tlA := topLeft.RGBA()
	trR, trG, trB, trA := topRight.RGBA()
	blR, blG, blB, blA := bottomLeft.RGBA()
	brR, brG, brB, brA := bottomRight.RGBA()

	interpolate := func(topLeft, topRight, bottomLeft, bottomRight uint32) byte {
		top := float64(topLeft)*(1-xWeight) + float64(topRight)*xWeight
		bottom := float64(bottomLeft)*(1-xWeight) + float64(bottomRight)*xWeight
		return byte(math.Round((top*(1-yWeight) + bottom*yWeight) / 257))
	}

	return interpolate(tlR, trR, blR, brR),
		interpolate(tlG, trG, blG, brG),
		interpolate(tlB, trB, blB, brB),
		interpolate(tlA, trA, blA, brA)
}

func loadBankLogo(bankBIN string) (image.Image, error) {
	logoLookupOnce.Do(loadLogoLookup)
	if logoLookupErr != nil {
		return nil, logoLookupErr
	}

	bankBIN = strings.TrimSpace(bankBIN)
	assetPath := logos.byBIN[bankBIN]
	if assetPath == "" {
		if bank, exists := VNBankM[bankBIN]; exists {
			assetPath = logos.byCode[normalizeLogoKey(bank.Code)]
			if assetPath == "" {
				assetPath = logos.byShortName[normalizeLogoKey(bank.ShortName)]
			}
		}
	}
	if assetPath == "" {
		return nil, fmt.Errorf("no embedded bank logo for BIN %q", bankBIN)
	}
	return decodePNGAsset(assetPath)
}

func loadLogoLookup() {
	data, err := pngAssets.ReadFile("image/banks/index.json")
	if err != nil {
		logoLookupErr = fmt.Errorf("read bank logo index: %w", err)
		return
	}
	var manifest logoManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		logoLookupErr = fmt.Errorf("decode bank logo index: %w", err)
		return
	}

	logos = logoLookup{
		byBIN:       make(map[string]string),
		byCode:      make(map[string]string),
		byShortName: make(map[string]string),
	}
	for _, entry := range manifest.Logos {
		assetPath := path.Join("image/banks", entry.File)
		if _, err := pngAssets.Open(assetPath); err != nil {
			logoLookupErr = fmt.Errorf("bank logo %q: %w", entry.File, err)
			return
		}
		for _, bin := range strings.Split(entry.BIN, ",") {
			if bin = strings.TrimSpace(bin); bin != "" {
				logos.byBIN[bin] = assetPath
			}
		}
		logos.byCode[normalizeLogoKey(entry.Code)] = assetPath
		logos.byShortName[normalizeLogoKey(entry.ShortName)] = assetPath
	}
}

func normalizeLogoKey(value string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
}

func decodePNGAsset(assetPath string) (image.Image, error) {
	data, err := pngAssets.ReadFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("read asset %q: %w", assetPath, err)
	}
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode asset %q: %w", assetPath, err)
	}
	return decoded, nil
}
