package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestMergeBanksKeepsLocalDataAndAddsMomoVietQRBINs(t *testing.T) {
	local := []bankRecord{
		{BIN: "970437", Code: "LOCAL-HDB", Name: "Local HDB"},
	}
	remote := map[string]momoBank{
		"HDB": {
			BIN:       "970437, 970420",
			ShortName: "HDBank",
			Name:      "Remote HDB",
			IsVietQR:  true,
		},
		"CAKE": {
			BIN:       "546034",
			ShortName: "CAKE",
			Name:      "Cake by VPBank",
			IsVietQR:  true,
		},
		"NOTVIETQR": {
			BIN:       "123456",
			ShortName: "Ignored",
			Name:      "Ignored",
			IsVietQR:  false,
		},
		"INVALID": {
			BIN:       "12345X",
			ShortName: "Invalid",
			Name:      "Invalid",
			IsVietQR:  true,
		},
	}

	got := mergeBanks(local, remote)
	want := []bankRecord{
		{BIN: "970437", Code: "LOCAL-HDB", Name: "Local HDB"},
		{BIN: "546034", Code: "CAKE", ShortName: "CAKE", Name: "Cake by VPBank"},
		{BIN: "970420", Code: "HDB", ShortName: "HDBank", Name: "Remote HDB"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeBanks() = %#v, want %#v", got, want)
	}
}

func TestFetchMomoBanks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"CAKE":{"bin":"546034","shortName":"CAKE","name":"Cake","isVietQr":true}}`))
	}))
	defer server.Close()

	banks, err := fetchMomoBanks(server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got := banks["CAKE"]; got.BIN != "546034" || !got.IsVietQR {
		t.Fatalf("CAKE = %#v", got)
	}
}

func TestFetchMomoBanksRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := fetchMomoBanks(server.Client(), server.URL)
	if err == nil || !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("error = %v, want HTTP 503", err)
	}
}
