package vietqr_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os/exec"
	"strings"
	"testing"

	"github.com/subiz/vietqr"
)

const cakePayload = "00020101021238540010A00000072701240006546034011003648218950208QRIBFTTA5303704540717000005802VN62240820SM83 PHAM KIEU THANH63040C94"

func TestRenderPNG(t *testing.T) {
	data, err := vietqr.RenderPNG(cakePayload, "546034")
	if err != nil {
		t.Fatal(err)
	}

	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if config.Width != vietqr.PNGWidth || config.Height != vietqr.PNGHeight {
		t.Fatalf("PNG size = %dx%d, want %dx%d", config.Width, config.Height, vietqr.PNGWidth, vietqr.PNGHeight)
	}

	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if !isWhite(decoded.At(10, 10)) {
		t.Error("outer margin is not white")
	}
	if isWhite(decoded.At(75, 190)) {
		t.Error("QR frame is missing")
	}
	assertRegionHasContent(t, decoded, image.Rect(175, 45, 675, 170), "VietQR logo")
	assertRegionHasContent(t, decoded, image.Rect(95, 210, 755, 870), "QR code")
	assertQRMargin(t, decoded, image.Rect(75, 190, 775, 890), 30)
	assertRegionHasContent(t, decoded, image.Rect(70, 950, 390, 1035), "NAPAS logo")
	assertRegionHasContent(t, decoded, image.Rect(465, 935, 785, 1045), "bank logo")
}

func TestGeneratePNG(t *testing.T) {
	data, err := vietqr.GeneratePNG(1700000, "546034", "0364821895", "SM83 PHAM KIEU THANH")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("GeneratePNG returned no data")
	}
}

func TestRenderPNGBankLogoVariants(t *testing.T) {
	for _, bankBIN := range []string{"970416", "970468", "546034"} {
		t.Run(bankBIN, func(t *testing.T) {
			data, err := vietqr.RenderPNG(cakePayload, bankBIN)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) == 0 {
				t.Fatal("RenderPNG returned no data")
			}
		})
	}
}

func TestRenderPNGUnknownBankLogo(t *testing.T) {
	_, err := vietqr.RenderPNG(cakePayload, "000000")
	if err == nil || !strings.Contains(err.Error(), "no embedded bank logo") {
		t.Fatalf("error = %v", err)
	}
}

func TestRenderPNGDecodesWithZBar(t *testing.T) {
	if _, err := exec.LookPath("zbarimg"); err != nil {
		t.Skip("zbarimg is not installed")
	}
	data, err := vietqr.RenderPNG(cakePayload, "546034")
	if err != nil {
		t.Fatal(err)
	}

	command := exec.Command("zbarimg", "--quiet", "-Sdisable", "-Sqrcode.enable", "-")
	command.Stdin = bytes.NewReader(data)
	output, err := command.Output()
	if err != nil {
		t.Fatalf("zbarimg: %v", err)
	}
	decoded := strings.TrimSuffix(strings.TrimPrefix(string(output), "QR-Code:"), "\n")
	if decoded != cakePayload {
		t.Fatalf("decoded payload = %q, want %q", decoded, cakePayload)
	}
}

func assertRegionHasContent(t *testing.T, source image.Image, region image.Rectangle, name string) {
	t.Helper()
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			if !isWhite(source.At(x, y)) {
				return
			}
		}
	}
	t.Errorf("%s region is blank", name)
}

func assertQRMargin(t *testing.T, source image.Image, frame image.Rectangle, maxMargin int) {
	t.Helper()
	blackBounds := image.Rectangle{}
	found := false
	for y := frame.Min.Y; y < frame.Max.Y; y++ {
		for x := frame.Min.X; x < frame.Max.X; x++ {
			if !isBlack(source.At(x, y)) {
				continue
			}
			pixel := image.Rect(x, y, x+1, y+1)
			if !found {
				blackBounds = pixel
				found = true
			} else {
				blackBounds = blackBounds.Union(pixel)
			}
		}
	}
	if !found {
		t.Fatal("QR code has no black pixels")
	}

	margins := []int{
		blackBounds.Min.X - frame.Min.X,
		blackBounds.Min.Y - frame.Min.Y,
		frame.Max.X - blackBounds.Max.X,
		frame.Max.Y - blackBounds.Max.Y,
	}
	for _, margin := range margins {
		if margin > maxMargin {
			t.Fatalf("QR margin = %dpx, want <= %dpx", margin, maxMargin)
		}
	}
}

func isWhite(value color.Color) bool {
	red, green, blue, alpha := value.RGBA()
	return alpha > 0xf000 && red > 0xf000 && green > 0xf000 && blue > 0xf000
}

func isBlack(value color.Color) bool {
	red, green, blue, alpha := value.RGBA()
	return alpha > 0xf000 && red < 0x2000 && green < 0x2000 && blue < 0x2000
}
