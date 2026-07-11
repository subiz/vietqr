package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	bankCSVPath      = "./bank.csv"
	generatedGoPath  = "./bank_generated.go"
	momoBankCodesURL = "https://payment.momo.vn/v2/gateway/api/bankcodes"
)

type bankRecord struct {
	BIN           string
	Name          string
	ShortName     string
	Code          string
	SWIFTCode     string
	AndroidBundle string
}

type momoBank struct {
	BIN       string `json:"bin"`
	ShortName string `json:"shortName"`
	Name      string `json:"name"`
	IsVietQR  bool   `json:"isVietQr"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	localBanks, err := readLocalBanks(bankCSVPath)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	momoBanks, err := fetchMomoBanks(client, momoBankCodesURL)
	if err != nil {
		return err
	}

	banks := mergeBanks(localBanks, momoBanks)
	source, err := generateSource(banks)
	if err != nil {
		return err
	}

	if err := os.WriteFile(generatedGoPath, source, 0600); err != nil {
		return fmt.Errorf("write %s: %w", generatedGoPath, err)
	}

	fmt.Printf("generated %s with %d local and %d total banks\n", generatedGoPath, len(localBanks), len(banks))
	return nil
}

func readLocalBanks(path string) ([]bankRecord, error) {
	bankFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer bankFile.Close()

	records, err := csv.NewReader(bankFile).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("read %s: no bank records", path)
	}

	banks := make([]bankRecord, 0, len(records)-1)
	seen := make(map[string]struct{}, len(records)-1)
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 7 {
			return nil, fmt.Errorf("read %s: row %d has %d columns, want at least 7", path, i+1, len(record))
		}

		bin := strings.TrimSpace(record[1])
		if !validBIN(bin) {
			return nil, fmt.Errorf("read %s: row %d has invalid BIN %q", path, i+1, bin)
		}
		if _, exists := seen[bin]; exists {
			return nil, fmt.Errorf("read %s: duplicate BIN %q", path, bin)
		}
		seen[bin] = struct{}{}

		banks = append(banks, bankRecord{
			BIN:           bin,
			Code:          strings.TrimSpace(record[2]),
			ShortName:     strings.TrimSpace(record[3]),
			SWIFTCode:     strings.TrimSpace(record[4]),
			AndroidBundle: strings.TrimSpace(record[5]),
			Name:          strings.TrimSpace(record[6]),
		})
	}

	return banks, nil
}

func fetchMomoBanks(client *http.Client, url string) (map[string]momoBank, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create MoMo bank request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "github.com/subiz/vietqr bank generator")

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
	hasVietQR := false
	for _, bank := range banks {
		if bank.IsVietQR {
			hasVietQR = true
			break
		}
	}
	if !hasVietQR {
		return nil, fmt.Errorf("decode MoMo banks: response has no VietQR banks")
	}

	return banks, nil
}

func mergeBanks(local []bankRecord, remote map[string]momoBank) []bankRecord {
	merged := append([]bankRecord(nil), local...)
	seen := make(map[string]struct{}, len(local))
	for _, bank := range local {
		seen[bank.BIN] = struct{}{}
	}

	codes := make([]string, 0, len(remote))
	for code := range remote {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		bank := remote[code]
		if !bank.IsVietQR {
			continue
		}

		for _, bin := range strings.Split(bank.BIN, ",") {
			bin = strings.TrimSpace(bin)
			if !validBIN(bin) {
				continue
			}
			if _, exists := seen[bin]; exists {
				continue
			}

			merged = append(merged, bankRecord{
				BIN:       bin,
				Name:      strings.TrimSpace(bank.Name),
				ShortName: strings.TrimSpace(bank.ShortName),
				Code:      strings.TrimSpace(code),
			})
			seen[bin] = struct{}{}
		}
	}

	return merged
}

func validBIN(bin string) bool {
	if len(bin) != 6 {
		return false
	}
	for _, char := range bin {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func generateSource(banks []bankRecord) ([]byte, error) {
	var gen bytes.Buffer
	gen.WriteString(`// CAUTION: THIS IS GENERATED CODE, DO NOT EDIT
// Local data from ./bank.csv is authoritative. Missing VietQR BINs are added
// from https://payment.momo.vn/v2/gateway/api/bankcodes.
// To regenerate this file, run "go run ./cmd/generate.go" from the repository root.
package vietqr

// VNBankM maps BIN to Bank information.
var VNBankM = map[string]Bank{
`)

	for _, bank := range banks {
		fmt.Fprintf(&gen, `	%q: Bank{
		BIN:           %q,
		Name:          %q,
		ShortName:     %q,
		Code:          %q,
		SWIFTCode:     %q,
		AndroidBundle: %q,
	},
`, bank.BIN, bank.BIN, bank.Name, bank.ShortName, bank.Code, bank.SWIFTCode, bank.AndroidBundle)
	}
	gen.WriteString("}\n")

	source, err := format.Source(gen.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}
	return source, nil
}
