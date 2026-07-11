package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadAll(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/banks":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]momoBank{
				"CAKE": {
					BIN:         "546034",
					ShortName:   "CAKE",
					Name:        "Cake by VPBank",
					BankLogoURL: server.URL + "/cake.png",
					IsVietQR:    true,
				},
				"OTHER": {
					BIN:         "123456",
					ShortName:   "Other",
					Name:        "Other Bank",
					BankLogoURL: server.URL + "/other.svg",
					IsVietQR:    false,
				},
			})
		case "/cake.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("png-data"))
		case "/other.svg":
			w.Header().Set("Content-Type", "image/svg+xml")
			_, _ = w.Write([]byte("<svg></svg>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	outputDir := t.TempDir()
	downloaded, err := downloadAll(server.Client(), server.URL+"/banks", outputDir, false)
	if err != nil {
		t.Fatal(err)
	}
	if downloaded != 2 {
		t.Fatalf("downloaded = %d, want 2", downloaded)
	}

	assertFileContents(t, filepath.Join(outputDir, "CAKE.png"), "png-data")
	assertFileContents(t, filepath.Join(outputDir, "OTHER.svg"), "<svg></svg>")

	manifestData, err := os.ReadFile(filepath.Join(outputDir, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got manifest
	if err := json.Unmarshal(manifestData, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Logos) != 2 || got.Logos[0].Code != "CAKE" || got.Logos[0].File != "CAKE.png" {
		t.Fatalf("manifest = %#v", got)
	}
}

func TestDownloadAllVietQROnly(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/banks" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]momoBank{
				"CAKE":  {BankLogoURL: server.URL + "/logo", IsVietQR: true},
				"OTHER": {BankLogoURL: server.URL + "/logo", IsVietQR: false},
			})
			return
		}
		w.Header().Set("Content-Type", "image/webp")
		_, _ = w.Write([]byte("image"))
	}))
	defer server.Close()

	downloaded, err := downloadAll(server.Client(), server.URL+"/banks", t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	if downloaded != 1 {
		t.Fatalf("downloaded = %d, want 1", downloaded)
	}
}

func TestSafeFileName(t *testing.T) {
	if got := safeFileName("../Bank Code"); got != "___Bank_Code" {
		t.Fatalf("safeFileName() = %q", got)
	}
}

func assertFileContents(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Fatalf("%s = %q, want %q", path, data, want)
	}
}
