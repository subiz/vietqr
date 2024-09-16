package vietqr_test

import (
	"fmt"
	"testing"

	"github.com/subiz/vietqr"
)

func TestQR(t *testing.T) {
	testcases := []struct {
		onetime       bool
		servicetype   string
		amount        int
		bankBIN       string
		accountnumber string
		note          string
		expect        string
	}{
		{false, "QRIBFTTA", 0, "970423", "0099999999", "",
			"00020101021138540010A00000072701240006970423011000999999990208QRIBFTTA53037045802VN6304CBB4"},

		{true, "QRIBFTTA", 40123, "970422", "0023457923442", "test text string",
			"00020101021238570010A00000072701270006970422011300234579234420208QRIBFTTA53037045405401235802VN62200816test text string6304D9C6"},

		{true, "QRIBFTTA", 40123, "970422", "0023457923442", "test text string",
			"00020101021238570010A00000072701270006970422011300234579234420208QRIBFTTA53037045405401235802VN62200816test text string6304D9C6"},

		{true, "QRIBFTTA", 40123, "970422", "0023457923442", "chuyển khoản",
			"00020101021238570010A00000072701270006970422011300234579234420208QRIBFTTA53037045405401235802VN62160812chuyen khoan6304722F"},

		{true, "QRIBFTTA", 40123, "970422", "0023457923442", "chuyển khoản",
			"00020101021238570010A00000072701270006970422011300234579234420208QRIBFTTA53037045405401235802VN62160812chuyen khoan6304722F"},
		{true, "QRIBFTTA", 40123, "970422", "0023457923442ASDFLJ", "chuyen khoan alsdkf laksjdflk asjdflja slkdalks djflkasjd fajsldk jalskdfj lkasjdflk ajslkfj l",
			"00020101021238630010A0000007270133000697042201190023457923442ASDFLJ0208QRIBFTTA53037045405401235802VN62290825chuyen khoan alsdkf laksj6304E5DB"},

		// account number and note too long -> must trim
		{true, "QRIBFTTA", 40123, "970422", "0023457923442ASDFLJ111111", "chuyen khoan alsdkf laksjdflk asjdflja slkdalks djflkasjd fajsldk jalskdfj lkasjdflk ajslkfj l",
			"00020101021238630010A0000007270133000697042201190023457923442ASDFLJ0208QRIBFTTA53037045405401235802VN62290825chuyen khoan alsdkf laksj6304E5DB"},
	}

	for i, tc := range testcases {
		out := vietqr.GenerateWithParams(tc.onetime, tc.servicetype, float64(tc.amount), tc.bankBIN, tc.accountnumber, tc.note, "VND", "")
		if out != tc.expect {
			t.Errorf("SHOULD EQ IN [%d], out [%s], expect [%s]", i, out, tc.expect)
		}
	}
}

func TestBank(t *testing.T) {
	testCases := []struct {
		BIN  string
		Code string
	}{
		{"970412", "PVCB"},
		{"970425", "ABB"},
	}

	if len(vietqr.VNBankM) != 56 {
		t.Errorf("Must have %d banks, but got %d banks", 56, len(vietqr.VNBankM))
	}
	for _, tc := range testCases {
		bank := vietqr.VNBankM[tc.BIN]
		if bank.Code != tc.Code {
			t.Errorf("SHOULD EQ for Bank BIN [%s], expect [%s], got [%s]", tc.BIN, tc.Code, bank.Code)
		}
	}
}

func ExampleString() {
	code := vietqr.Generate(120000, "970415", "0011001932418", "ủng hộ lũ lụt")
	fmt.Println(code)
	// Output: 00020101021238570010A00000072701270006970415011300110019324180208QRIBFTTA530370454061200005802VN62170813ung ho lu lut6304C15C
}
