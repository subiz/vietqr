// Derived from github.com/skip2/go-qrcode.
// Copyright 2014 Tom Harwood. Licensed under the MIT license in LICENSE.

package qrcode

import "testing"

func TestQRCodeISOAnnexIExample(t *testing.T) {
	code, err := New("01234567", Medium)
	if err != nil {
		t.Fatal(err)
	}
	code.encode()
	if code.mask != 2 {
		t.Fatalf("ISO Annex I mask = %d, want 2", code.mask)
	}
}
