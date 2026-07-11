package vietqr_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/subiz/vietqr"
)

func TestCrc(t *testing.T) {
	testcases := []struct {
		in     string
		expect string
	}{
		{"00", "2EC9"},
		{"01", "3EE8"},
		{"00020101021138600010A00000072701300006970403011697040311012345670208QRIBFTTC53037045802VN6304", "4F52"},
		{"00020101021238570010A00000072701270006970403011300110123456780208QRIBFTTA530370454061800005802VN62340107NPS68690819thanh toan don hang6304", "2E2E"},
		{"00020101021238600010A00000072701300006970403011697040311012345670208QRIBFTTC530370454061800005802VN62340107NPS68690819thanh toan don hang6304", "A203"},
		{"00020101021138540010A00000072701240006970423011000999999990208QRIBFTTA53037045802VN6304", "CBB4"},
	}

	for _, tc := range testcases {
		out := vietqr.CrcChecksum(tc.in)
		if out != tc.expect {
			t.Errorf("SHOULD EQ IN [%s], out [%s], expect [%s]", tc.in, out, tc.expect)
		}
	}
}

func TestConvertASCII(t *testing.T) {
	testcases := []struct {
		in     string
		expect string
	}{
		{"ậậậậ", "aaaa"},
		{"Bằng Minh Tuấn", "Bang Minh Tuan"},
		{"ß", ""},
		{"한글", ""},
		{"æ", ""},
		{"イーブイ", ""},
		{"Cộng hòa xã hội chủ nghĩa Việt Nam. Độc lập tự do - hạnh phúc", "Cong hoa xa hoi chu nghia Viet Nam. Doc lap tu do - hanh phuc"},
		{"République socialiste du Vietnam. Indépendance et liberté - bonheur", "Republique socialiste du Vietnam. Independance et liberte - bonheur"},
		{"Vietnam Sosyalist Cumhuriyeti. Bağımsızlık ve özgürlük - mutluluk", "Vietnam Sosyalist Cumhuriyeti. Bamszlk ve zgrlk - mutluluk"},
		{"Социјалистичке Републике Вијетнам. Независност и слобода - срећа", "  .    - "},
		{"越南社会主义共和国。独立与自由——幸福", ""},
	}

	for _, tc := range testcases {
		out := vietqr.Ascii(tc.in)
		if out != tc.expect {
			t.Errorf("expect \"%s\", got \"%s\"", tc.expect, out)
		}
	}
}

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
		{true, "QRIBFTTA", 1700000, "546034", "0364821895", "SM83 PHAM KIEU THANH",
			"00020101021238540010A00000072701240006546034011003648218950208QRIBFTTA5303704540717000005802VN62240820SM83 PHAM KIEU THANH63040C94"},

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

func TestGenerateWithParamsFormatsAmountAndCountry(t *testing.T) {
	testcases := []struct {
		currency      string
		amount        float64
		country       string
		expectAmount  string
		expectCountry string
	}{
		{"VND", 1700000, "VN", "1700000", "VN"},
		{"JPY", 123.6, "JP", "124", "JP"},
		{"KRW", 123.4, "KR", "123", "KR"},
		{"MYR", 123.4, "MY", "123.40", "MY"},
		{"CNY", 123.456, "RC", "123.46", "RC"},
		{"IDR", 123.4, "RI", "123.40", "RI"},
		{"PHP", 123.4, "RP", "123.40", "RP"},
		{"SGD", 123.4, "SG", "123.40", "SG"},
		{"THB", 123.4, "TH", "123.40", "TH"},
		{"VND", 9999999999999, "VN", "9999999999999", "VN"},
		{"MYR", 9999999999.99, "MY", "9999999999.99", "MY"},
		{"VND", 1, "invalid", "1", "VN"},
	}

	for _, tc := range testcases {
		t.Run(tc.currency+"_"+tc.country, func(t *testing.T) {
			out := vietqr.GenerateWithParams(true, "QRIBFTTA", tc.amount, "970422", "0023457923442", "", tc.currency, tc.country)
			if amount, found := rootTLVValue(out, "54"); !found || amount != tc.expectAmount {
				t.Errorf("amount = %q, found = %v; want %q", amount, found, tc.expectAmount)
			}
			if country, found := rootTLVValue(out, "58"); !found || country != tc.expectCountry {
				t.Errorf("country = %q, found = %v; want %q", country, found, tc.expectCountry)
			}
		})
	}
}

func TestGenerateWithParamsOmitsInvalidAmount(t *testing.T) {
	testcases := []struct {
		name     string
		currency string
		amount   float64
	}{
		{"too long", "VND", 10000000000000},
		{"rounds to zero", "MYR", 0.001},
		{"not a number", "VND", math.NaN()},
		{"infinity", "VND", math.Inf(1)},
		{"unsupported currency", "USD", 10},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			out := vietqr.GenerateWithParams(true, "QRIBFTTA", tc.amount, "970422", "0023457923442", "", tc.currency, "VN")
			if amount, found := rootTLVValue(out, "54"); found {
				t.Errorf("unexpected amount %q", amount)
			}
		})
	}
}

func TestPostalCodeDefinition(t *testing.T) {
	for _, def := range vietqr.Defaults {
		if strings.Contains(def.Name, "Postal Code") {
			if def.ID != "61" {
				t.Fatalf("Postal Code ID = %q, want 61", def.ID)
			}
			return
		}
	}
	t.Fatal("Postal Code definition not found")
}

func rootTLVValue(payload, wantedID string) (string, bool) {
	for offset := 0; offset+4 <= len(payload); {
		id := payload[offset : offset+2]
		length, err := strconv.Atoi(payload[offset+2 : offset+4])
		if err != nil || offset+4+length > len(payload) {
			return "", false
		}
		value := payload[offset+4 : offset+4+length]
		if id == wantedID {
			return value, true
		}
		offset += 4 + length
	}
	return "", false
}

func TestBank(t *testing.T) {
	testCases := []struct {
		BIN       string
		Code      string
		SWIFTCode string
	}{
		{"970412", "PVCB", "WBVNVNVX"},
		{"970425", "ABB", "ABBKVNVX"},
		{"970422", "MB", "MSCBVNVX"},
		{"546034", "CAKE", ""},
	}

	if len(vietqr.VNBankM) < 56 {
		t.Errorf("Must have at least %d banks, but got %d banks", 56, len(vietqr.VNBankM))
	}
	for _, tc := range testCases {
		bank := vietqr.VNBankM[tc.BIN]
		if bank.Code != tc.Code {
			t.Errorf("SHOULD EQ for Bank BIN [%s], expect [%s], got [%s]", tc.BIN, tc.Code, bank.Code)
		}

		if bank.SWIFTCode != tc.SWIFTCode {
			t.Errorf("SWIFT code should equal for Bank BIN [%s], expect [%s], got [%s]", tc.BIN, tc.SWIFTCode, bank.SWIFTCode)
		}
	}
}

func ExampleGenerate() {
	code := vietqr.Generate(120000, "970415", "0011001932418", "ủng hộ lũ lụt")
	fmt.Println(code)
	// Output: 00020101021238570010A00000072701270006970415011300110019324180208QRIBFTTA530370454061200005802VN62170813ung ho lu lut6304C15C
}
