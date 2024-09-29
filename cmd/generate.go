package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func main() {
	bankFile, err := os.Open("./bank.csv")
	if err != nil {
		panic(err)
	}
	defer bankFile.Close()

	csvReader := csv.NewReader(bankFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	gen := `// CAUTION: THIS IS GENERATED CODE, DO NOT EDIT
// The data is taken from ./bank.csv
// To regenerated this file: in this directory, run this command "go run ./cmd/bank.go"
package vietqr

// VNBankM maps BIN to Bank information
var VNBankM = map[string]Bank{`
	for i, record := range records {
		if i == 0 { // skip header
			continue
		}

		gen += fmt.Sprintf(`
	"%s": Bank{
		BIN:           "%s",
		Name:          "%s",
		ShortName:     "%s",
		Code:          "%s",
		SWIFTCode:     "%s",
		AndroidBundle: "%s",
	},`, strings.TrimSpace(record[1]), strings.TrimSpace(record[1]), strings.TrimSpace(record[6]), strings.TrimSpace(record[3]), strings.TrimSpace(record[2]), strings.TrimSpace(record[4]), strings.TrimSpace(record[5]))
	}

	gen += `
}
`
	if err := os.WriteFile("./bank_generated.go", []byte(gen), 0600); err != nil {
		panic(err)
	}
}
