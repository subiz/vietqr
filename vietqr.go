package vietqr

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const CRC_INIT uint16 = 0xFFFF
const CRC_POLY uint16 = 0x1021

var VNMAP = map[rune]rune{
	'ạ': 'a', 'ả': 'a', 'ã': 'a', 'à': 'a', 'á': 'a', 'â': 'a', 'ậ': 'a', 'ầ': 'a', 'ấ': 'a',
	'ẩ': 'a', 'ẫ': 'a', 'ă': 'a', 'ắ': 'a', 'ằ': 'a', 'ặ': 'a', 'ẳ': 'a', 'ẵ': 'a',
	'ó': 'o', 'ò': 'o', 'ọ': 'o', 'õ': 'o', 'ỏ': 'o', 'ô': 'o', 'ộ': 'o', 'ổ': 'o', 'ỗ': 'o',
	'ồ': 'o', 'ố': 'o', 'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ợ': 'o', 'ở': 'o', 'ỡ': 'o',
	'é': 'e', 'è': 'e', 'ẻ': 'e', 'ẹ': 'e', 'ẽ': 'e', 'ê': 'e', 'ế': 'e', 'ề': 'e', 'ệ': 'e', 'ể': 'e', 'ễ': 'e',
	'ú': 'u', 'ù': 'u', 'ụ': 'u', 'ủ': 'u', 'ũ': 'u', 'ư': 'u', 'ự': 'u', 'ữ': 'u', 'ử': 'u', 'ừ': 'u', 'ứ': 'u',
	'í': 'i', 'ì': 'i', 'ị': 'i', 'ỉ': 'i', 'ĩ': 'i',
	'ý': 'y', 'ỳ': 'y', 'ỷ': 'y', 'ỵ': 'y', 'ỹ': 'y',
	'đ': 'd',
	'Ạ': 'A', 'Ả': 'A', 'Ã': 'A', 'À': 'A', 'Á': 'A', 'Â': 'A', 'Ậ': 'A', 'Ầ': 'A', 'Ấ': 'A',
	'Ẩ': 'A', 'Ẫ': 'A', 'Ă': 'A', 'Ắ': 'A', 'Ằ': 'A', 'Ặ': 'A', 'Ẳ': 'A', 'Ẵ': 'A',
	'Ó': 'O', 'Ò': 'O', 'Ọ': 'O', 'Õ': 'O', 'Ỏ': 'O', 'Ô': 'O', 'Ộ': 'O', 'Ổ': 'O', 'Ỗ': 'O',
	'Ồ': 'O', 'Ố': 'O', 'Ơ': 'O', 'Ờ': 'O', 'Ớ': 'O', 'Ợ': 'O', 'Ở': 'O', 'Ỡ': 'O',
	'É': 'E', 'È': 'E', 'Ẻ': 'E', 'Ẹ': 'E', 'Ẽ': 'E', 'Ê': 'E', 'Ế': 'E', 'Ề': 'E', 'Ệ': 'E', 'Ể': 'E', 'Ễ': 'E',
	'Ú': 'U', 'Ù': 'U', 'Ụ': 'U', 'Ủ': 'U', 'Ũ': 'U', 'Ư': 'U', 'Ự': 'U', 'Ữ': 'U', 'Ử': 'U', 'Ừ': 'U', 'Ứ': 'U',
	'Í': 'I', 'Ì': 'I', 'Ị': 'I', 'Ỉ': 'I', 'Ĩ': 'I',
	'Ý': 'Y', 'Ỳ': 'Y', 'Ỷ': 'Y', 'Ỵ': 'Y', 'Ỹ': 'Y',
	'Đ': 'D',
}

var ISO_IEC_13239_data [256]uint16

func initCrcTable(poly uint16) {
	for n := uint16(0); n < 256; n++ {
		crc := n << 8
		for i := 0; i < 8; i++ {
			bit := (crc & 0x8000) != 0
			crc <<= 1
			if bit {
				crc ^= poly
			}
		}
		ISO_IEC_13239_data[n] = crc
	}
}

const OPTIONAL = "O"
const CONDITIONALLY = "C"
const MANDATORY = "M"

type Object struct {
	Name     string
	ID       string // 2 character
	Len      int
	MaxLen   int    // only used when Len is 0
	Required string // O C M

	Sub []Object
}

var MerchantAccount = []Object{
	{"Định danh duy nhất toàn cầu GUID ", "00", 0, 32, MANDATORY, nil}, // A000000727.
	{"Tổ chức thụ hưởng (NHTV, TGTT)      ", "01", 0, 1000, MANDATORY, MerchantAccountBeneficiaryOrganization},

	// QRIBFTTC: Mã dịch vụ chuyển tiền nhanh 24/7 bằng QR đến thẻ;
	// QRIBFTTA: Mã dịch vụ chuyển tiền nhanh 24/7 bằng QR đến tài khoản
	{"Mã dịch vụ                       ", "02", 0, 10, CONDITIONALLY, nil},
}

var MerchantAccountBeneficiaryOrganization = []Object{
	{"Acquier ID/BNB ID                ", "00", 6, 0, MANDATORY, nil},
	{"Merchant ID/Consumer ID          ", "01", 0, 19, MANDATORY, nil},
}

var AdditionalDataFieldTemplate = []Object{
	// Số hóa đơn/biên lai cung cấp bởi Merchant hoặc do KH tự nhập vào app.
	{"Bill Number                      ", "01", 0, 25, OPTIONAL, nil},

	// Số điện thoại di động có thể do merchant cung cấp hoặc do khách hàng tự nhập.
	{"Mobile Number                    ", "02", 0, 25, OPTIONAL, nil},
	{"Store Label                      ", "03", 0, 25, OPTIONAL, nil},
	{"Loyalty Number                   ", "04", 0, 25, OPTIONAL, nil},
	{"Reference Label                  ", "05", 0, 25, OPTIONAL, nil},
	{"Customer Label                   ", "06", 0, 25, OPTIONAL, nil},
	{"Terminal Label                   ", "07", 0, 25, OPTIONAL, nil},
	{"Purpose of Transaction           ", "08", 0, 25, OPTIONAL, nil},

	// Yêu cầu dữ liệu KH bổ sung
	// Một hoặc nhiều ký tự sau đâycó thể xuất hiện, cho biết dữ liệu tương ứng cần được cung cấp trong quá trình khởi tạo giao dịch:
	// • "A" = Địa chỉ khách hàng
	// • "M" = SĐT khách hàng
	// • "E" = Địa chỉ email của khách hàng
	{"Additional Consumer Data Request ", "09", 0, 3, OPTIONAL, nil},
	{"RFU for EMVCo                    ", "10", 0, 25, OPTIONAL, nil},
	{"Payment System specific templates", "50", 0, 25, OPTIONAL, nil},
}

var Defaults = []Object{
	{"Payload Format Indicator         ", "00", 2, 0, MANDATORY, nil},
	{"Point of Initiation Method       ", "01", 2, 0, MANDATORY, nil},
	{"Merchant Account Information     ", "38", 0, 99, MANDATORY, MerchantAccount}, //A000000727.
	{"Merchant Category Code           ", "52", 4, 0, OPTIONAL, nil},
	{"Transaction Currency             ", "53", 3, 3, MANDATORY, nil},
	{"Transaction Amount               ", "54", 0, 13, CONDITIONALLY, nil},
	{"Tip or Convenience Indicator     ", "55", 2, 0, OPTIONAL, nil},
	{"Value of Convenience Fee Fixed   ", "56", 0, 13, CONDITIONALLY, nil},
	{"Value of Convenience Fee         ", "57", 0, 5, CONDITIONALLY, nil},
	{"Country Code                     ", "58", 2, 0, MANDATORY, nil},
	{"Merchant Name                    ", "59", 0, 25, OPTIONAL, nil},
	{"Merchant City                    ", "60", 0, 15, OPTIONAL, nil},
	{"Postal Code                      ", "51", 0, 10, OPTIONAL, nil},
	{"Additional Data Field Template   ", "62", 0, 99, CONDITIONALLY, AdditionalDataFieldTemplate},
	{"Merchant Information             ", "64", 0, 99, OPTIONAL, nil},
	// {"CRC                              ", "63",  4, 0, MANDATORY, nil},
}

// servicetype:
//
//	"QRIBFTTC": dịch vụ chuyển tiền nhanh 24/7 bằng QR đến thẻ
//	"QRIBFTTA": dịch vụ chuyển tiền nhanh 24/7 bằng QR đến tài khoản
func GenerateWithParams(onetime bool, servicetype string, amount int, bankBIN string, accountnumber, note string) string {
	contents := map[string]string{}
	contents["00"] = "01"

	// Point of Initiation Method
	if onetime {
		contents["01"] = "12" // QR động – áp dụng khi mã QR chỉ cho phép thực hiện một lần giao dịch.
	} else {
		contents["01"] = "11" // QR tĩnh – áp dụng khi mã QR cho phép thực hiện nhiều lần giao dịch.
	}

	contents["3800"] = "A000000727"
	contents["380100"] = bankBIN       // "970468"        // bnb id
	contents["380101"] = accountnumber // "0011009950446" // Consumer id
	contents["3802"] = servicetype
	contents["53"] = "704" // vnd
	if amount > 0 {
		contents["54"] = strconv.Itoa(amount)
	}

	note = strings.TrimSpace(note)
	if note != "" {
		contents["6208"] = note
	}
	contents["58"] = "VN" // JP KR MY RC RI RP SG TH

	// generate qr code
	out := generateObject(nil, "", "", contents) + "6304" // ID for crc
	return out + CrcChecksum(out)
}

func generateObject(defs []Object, prefixid, id string, contents map[string]string) string {
	var def Object
	for _, d := range defs {
		if id == d.ID {
			def = d
			break
		}
	}
	if len(defs) == 0 {
		def = Object{ID: "--", Sub: Defaults}
	}

	if def.ID == "" {
		return ""
	}

	// compound object: object contains sub objects
	if len(def.Sub) > 0 {
		content := ""
		for _, sub := range def.Sub {
			content += generateObject(def.Sub, prefixid+id, sub.ID, contents)
		}
		if def.ID == "--" {
			return content
		}
		if content != "" {
			return fmt.Sprintf("%s%02d%s", def.ID, len(content), content)
		}
		return ""
	}

	content := ascii(contents[prefixid+id])
	if content == "" {
		return ""
	}
	var length = len(content)
	if def.MaxLen > 0 {
		if len(content) > def.MaxLen {
			length = def.MaxLen
		}
	}
	if length > 99 {
		length = 99
	}
	content = Substring(content, length)
	return fmt.Sprintf("%s%02d%s", id, length, content)
}

func init() {
	initCrcTable(CRC_POLY)
}

// Checksum returns CRC checksum of data using scpecified algorithm represented by the Table.
func CrcChecksum(str string) string {
	data := []byte(str)
	crc := CRC_INIT
	for _, d := range data {
		crc = crc<<8 ^ ISO_IEC_13239_data[byte(crc>>8)^d]
	}

	return fmt.Sprintf("%X", crc)
}

func Substring(s string, end int) string {
	if s == "" {
		return ""
	}

	if end >= len(s) {
		return s
	}

	start_str_idx := 0
	i := 0
	for j := range s {
		if i == end {
			return s[start_str_idx:j]
		}
		i++
	}
	return s[start_str_idx:]
}

// Convert replaces all non-ascii characters to equivalent ascii characters
// e.g: â => a, đ => d, ...
// To ensure the output string is pure ascii, this function remove all
// characters that does not have equivalent ascii character, for example: 主
func ascii(text string) string {
	return strings.Map(func(r rune) rune {
		if r <= unicode.MaxASCII {
			return r
		}

		// fast case: only ascii or vietnamese
		if vnr := VNMAP[r]; vnr != 0 {
			return vnr
		}

		// remove all non-ascii
		return -1
	}, text)
}

// Sinh mã VietQR tới STK
// Generate(180000, "970423", "00134234", "ghi chu")
// Xem bank.csv để tra mã BIN của ngân hàng
func Generate(amount int, bankBIN string, accountnumber, note string) string {
	return GenerateWithParams(true, "QRIBFTTA", amount, bankBIN, accountnumber, note)
}
