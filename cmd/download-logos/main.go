package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultMomoBankCodesURL = "https://payment.momo.vn/v2/gateway/api/bankcodes"
	defaultOutputDir        = "./image/banks"
	maxLogoBytes            = 10 << 20
)

type momoBank struct {
	BIN         string `json:"bin"`
	ShortName   string `json:"shortName"`
	Name        string `json:"name"`
	BankLogoURL string `json:"bankLogoUrl"`
	IsVietQR    bool   `json:"isVietQr"`
}

type manifest struct {
	Source string          `json:"source"`
	Logos  []manifestEntry `json:"logos"`
}

type manifestEntry struct {
	Code      string `json:"code"`
	BIN       string `json:"bin"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	IsVietQR  bool   `json:"isVietQr"`
	SourceURL string `json:"sourceUrl"`
	File      string `json:"file"`
}

func main() {
	outputDir := flag.String("output", defaultOutputDir, "directory for downloaded logos and index.json")
	vietQROnly := flag.Bool("vietqr-only", false, "download only entries marked as supporting VietQR")
	flag.Parse()

	client := &http.Client{Timeout: 20 * time.Second}
	downloaded, err := downloadAll(client, defaultMomoBankCodesURL, *outputDir, *vietQROnly)
	if err != nil {
		log.Fatalf("downloaded %d logos with errors: %v", downloaded, err)
	}
	log.Printf("downloaded %d logos to %s", downloaded, *outputDir)
}

func downloadAll(client *http.Client, apiURL, outputDir string, vietQROnly bool) (int, error) {
	banks, err := fetchBanks(client, apiURL)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 0, fmt.Errorf("create output directory: %w", err)
	}

	codes := make([]string, 0, len(banks))
	for code := range banks {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	entries := make([]manifestEntry, 0, len(codes))
	var downloadErrors []error
	seenFiles := make(map[string]string, len(codes))
	for _, code := range codes {
		bank := banks[code]
		if vietQROnly && !bank.IsVietQR {
			continue
		}
		if strings.TrimSpace(bank.BankLogoURL) == "" {
			downloadErrors = append(downloadErrors, fmt.Errorf("%s: missing bankLogoUrl", code))
			continue
		}

		baseName := safeFileName(code)
		if previousCode, exists := seenFiles[baseName]; exists {
			downloadErrors = append(downloadErrors, fmt.Errorf("%s and %s produce the same filename", previousCode, code))
			continue
		}
		seenFiles[baseName] = code

		fileName, err := downloadLogo(client, bank.BankLogoURL, outputDir, baseName)
		if err != nil {
			downloadErrors = append(downloadErrors, fmt.Errorf("%s: %w", code, err))
			continue
		}

		entries = append(entries, manifestEntry{
			Code:      code,
			BIN:       bank.BIN,
			Name:      bank.Name,
			ShortName: bank.ShortName,
			IsVietQR:  bank.IsVietQR,
			SourceURL: bank.BankLogoURL,
			File:      fileName,
		})
	}

	manifestData, err := json.MarshalIndent(manifest{Source: apiURL, Logos: entries}, "", "  ")
	if err != nil {
		return len(entries), fmt.Errorf("encode logo manifest: %w", err)
	}
	manifestData = append(manifestData, '\n')
	if err := writeFileAtomic(filepath.Join(outputDir, "index.json"), manifestData, 0644); err != nil {
		return len(entries), fmt.Errorf("write logo manifest: %w", err)
	}

	return len(entries), errors.Join(downloadErrors...)
}

func fetchBanks(client *http.Client, apiURL string) (map[string]momoBank, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create MoMo bank request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "github.com/subiz/vietqr logo downloader")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch MoMo banks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("fetch MoMo banks: HTTP %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var banks map[string]momoBank
	if err := json.NewDecoder(resp.Body).Decode(&banks); err != nil {
		return nil, fmt.Errorf("decode MoMo banks: %w", err)
	}
	if len(banks) == 0 {
		return nil, fmt.Errorf("decode MoMo banks: empty response")
	}
	return banks, nil
}

func downloadLogo(client *http.Client, sourceURL, outputDir, baseName string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("create logo request: %w", err)
	}
	req.Header.Set("Accept", "image/*")
	req.Header.Set("User-Agent", "github.com/subiz/vietqr logo downloader")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch logo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch logo: HTTP %s", resp.Status)
	}

	contentType := strings.ToLower(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]))
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("fetch logo: unexpected Content-Type %q", contentType)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxLogoBytes+1))
	if err != nil {
		return "", fmt.Errorf("read logo: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("read logo: empty response")
	}
	if len(data) > maxLogoBytes {
		return "", fmt.Errorf("read logo: exceeds %d bytes", maxLogoBytes)
	}

	extension := logoExtension(contentType, sourceURL)
	fileName := baseName + extension
	if err := writeFileAtomic(filepath.Join(outputDir, fileName), data, 0644); err != nil {
		return "", fmt.Errorf("write logo: %w", err)
	}
	return fileName, nil
}

func logoExtension(contentType, sourceURL string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/gif":
		return ".gif"
	}

	if extensions, err := mime.ExtensionsByType(contentType); err == nil && len(extensions) > 0 {
		return extensions[0]
	}
	if parsed, err := url.Parse(sourceURL); err == nil {
		if extension := strings.ToLower(filepath.Ext(parsed.Path)); extension != "" {
			return extension
		}
	}
	return ".img"
}

func safeFileName(value string) string {
	value = strings.TrimSpace(value)
	return strings.Map(func(char rune) rune {
		switch {
		case char >= 'a' && char <= 'z':
			return char
		case char >= 'A' && char <= 'Z':
			return char
		case char >= '0' && char <= '9':
			return char
		case char == '-' || char == '_':
			return char
		default:
			return '_'
		}
	}, value)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".download-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(perm); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
