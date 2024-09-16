# VietQR
# [![GoDoc](https://godoc.org/github.com/subiz/vietqr?status.svg)](http://godoc.org/github.com/subiz/vietqr)

Sinh mã [VietQR](https://vietqr.net/) cho giao dịch chuyển khoản

![](./image/vietqr.png)


## Cài đặt
```
  go get -u github.com/subiz/vietqr
```

## Sử dụng

```go

import (
  "github.com/subiz/vietqr"
)

func main() {
  // Sinh mã chuyển khoản 120.000 vào STK 0011001932418 ngân hàng Vietin với ghi chú "ủng hộ lũ lụt"
  code := vietqr.Generate(120000, "970415", "0011001932418", "ủng hộ lũ lụt")
  fmt.Println(code)
  // 00020101021238570010A00000072701270006970415011300110019324180208QRIBFTTA530370454061200005802VN62170813ung ho lu lut6304C15C
}
```
### Mô tả cách sinh mã

VietQR là tiêu chuẩn QR thanh toán được phát hành bởi Napas và các ngân hàng thành viên. Website chính thức của VietQR là [vietqr.net](https://vietqr.net). Bạn có thể tìm thấy tài liệu mô tả chi tiết về tiêu chuẩn VietQR ở đó. Tôi cũng đã lưu trữ lại bản gốc trong thư mục `/spec`. Tài liệu gốc có thể sẽ hơi khó hiểu nên tôi tóm tắt lại ý chính giúp bạn dễ tiêu hóa hơn.

QR code là một hình ảnh được sinh ra từ một đoạn text. Bạn có thể dùng lệnh `qrencode` để sinh QR code cho chữ `Hello World!` như sau:

```sh
$ qrencode -t ASCII  'Hello World!'

        ##############    ##    ##  ##############
        ##          ##  ##    ##    ##          ##
        ##  ######  ##    ##        ##  ######  ##
        ##  ######  ##  ##    ##    ##  ######  ##
        ##  ######  ##      ######  ##  ######  ##
        ##          ##  ######  ##  ##          ##
        ##############  ##  ##  ##  ##############
                            ######
        ##########  ########    ####  ##  ##  ##
        ##########    ##  ##    ##############  ##
        ##    ##    ##  ##    ##  ####    ######
                ##      ##  ##    ##    ######
        ####  ##########  ######  ####          ##
                        ##  ##  ##        ##
        ##############  ####    ##  ##      ####
        ##          ##    ##########  ##  ####
        ##  ######  ##  ##      ##  ####      ####
        ##  ######  ##  ####      ##########
        ##  ######  ##  ######  ##    ##    ##
        ##          ##  ####    ####    ######
        ##############  ##########  ##  ##    ##

```


VietQR code được sinh ra từ một đoạn văn bản trông như này
```
00020101021138540010A00000072701240006970423011000999999990208QRIBFTTA53037045802VN6304CBB4
```

Một ví dụ khác:

```
00020101021238630010A0000007270133000697042201190023457923442ASDFLJ0208QRIBFTTA53037045405401235802VN62290825chuyen khoan alsdkf laksj6304E5DB
```

Mã VietQR được cấu thành bởi các đoạn text nhỏ hơn - "đối tượng dữ liệu". Đối tượng dữ liệu dùng để mô tả một thông tin trong giao dịch. Ví dụ số tiền cần chuyển khoản, mã ngân hàng, ...

Ví dụ:

| Đối tượng dữ liệu    | Ý nghĩa                                |
|----------------------|--------------------------------------- |
| `540540999`          | Số tiền chuyển khoản `40999`           |
| `0814ung ho bao lut` | Nội dung chyển khoản `ung ho bao lut`  |

Đối tượng dữ liệu *luôn* bao gồm 3 thành phần:
1. 2 ký tự đầu: ID của đối-tượng-dữ-liệu, nó cho biết đối tượng thể hiện thông tin gì.
2. 2 ký tự tiếp: Số lượng ký tự của thông tin.
3. Phần còn lại là thông tin dưới dạng unicode text.

```
0814ung ho bao lut
```

* ID: `08` (nội dung chuyển khoản)
* Độ dài dữ liệu: `14`
* Dữ liệu: `ung ho bao lut`

Đây là một số ID dùng trong mã VietQR

| ID    | Ý nghĩa                                |
|-------|--------------------------------------- |
| `38`    | Thông tin tài khoản thụ hưởng          |
| `53`    | Tiền tệ (VND là: `704`)                |
| `54`    | Số tiền                                |
| `58`    | Mã quốc gia                            |


Một số đối-tượng-dữ-liệu lại được cấu thành bởi các đối-tượng-dữ-liệu nhỏ hơn. Ví dụ "thông tin tài khoản thụ hưởng" ID 38:

`38` `57` `0010A00000072701270006970422011300234579234420208QRIBFTTA`

Các đối tượng dữ liệu con được tổ chức tương tự: `ID``Độ dài``Dữ liệu`. Với ví dụ ở trên, phần dữ liệu `0010A00000072701270006970422011300234579234420208QRIBFTTA` sẽ được dịch là

`00` `10` `A000000727` `01` `27` `000697042201130023457923442` `02` `08` `QRIBFTTA`

* Đối tượng dữ liệu đầu tiên là ID `00`, độ dài `10`, thông tin là `A000000727`
* Đối tượng dữ liệu tiếp theo là ID `01`, độ dài `27`, thông tin là `000697042201130023457923442`
* Đối tượng cuối cùng ID `02`, độ dài `08`, thông tin là `QRIBFTTA`

Mã VietQR ứng với giao dịch như sau:

* Ngân hàng `MBBank`
* STK thụ hưởng `002345792344`
* Số tiền `42123`
* Nội dung chuyển khoản: `ung ho bao lut`

Sẽ được sinh bằng các bước dưới đây:

**Bước 1**: Ghi lại Payload Format Indicator - phiên bản dữ liệu (ID 0)

`00` `02` `01`

Bạn sẽ để ý thấy rằng mã VietQR nào cũng bắt đầu bằng chuỗi `000201`, ý nghĩa của nó là VietQR code version 1

**Bước 2**: Thêm phương-thức-khởi-tạo (ID 1), nếu mã VietQR được dùng lại nhiều lần thì mang giá trị `11`, còn muốn quét 1 lần rồi vô hiệu thì dùng `12`

`01` `02` `11`

**Bước 3**: Thêm thông-tin-người-thụ-hưởng ID 38. Đây là một đối-tượng-dữ-liệu bao gồm 3 đối-tượng-dữ-liệu con. Bạn cần sinh phần con trước. Đầu tiên cần xác định mã BIN của ngân hàng MBbank là `970422` -> ghi số này vào ID 00: `0006970422`. Tiếp theo ghi STK thụ hưởng (ID 01): `0112002345792344`. Ghi tiếp chuỗi `0208QRIBFTTA` (ID 02 quy định việc chuyển vào thẻ hay TK ngân hàng). Cuối cùng đếm số ký tự và bọc lại trong ID 38

`38` `38` `000697042201120023457923440208QRIBFTTA`


**Bước 4**: thêm mã tiền tệ (ID 53): `704` (VND)

```
5303704
```

**Bước 5**: Thêm số tiền giao dịch (ID 54), viết liền không cách, ví dụ: `18.000`

```
5406180000
```

**Bước 6**: Thêm mã quốc gia ID 58

```
5802VN
```

**Bước 7**: Thêm ghi chú, ghi chú là thông tin con ID 08 của thông "tin thông tin bổ xung" ID 62. Viết ghi chú trước `0814ung ho bao lut`, sau đó đếm độ dài rồi bọc trong ID 01

```
62160814ung ho bao lut
```

**Bước 8**: Thêm mã CRC checksum ID 63

```
6304E69F
```

Cuối cùng ta được

```
0002010102113838000697042201120023457923440208QRIBFTTA530370454061800005802VN62160814ung ho bao lut6304E69F`
```

### Thông tin chiếu

1. Tra cứu mã BIN
https://www.sbv.gov.vn/webcenter/portal/vi/menu/trangchu/ttvnq/htmtcqht?_afrLoop=1982850809377774#%40%3F_afrLoop%3D1982850809377774%26centerWidth%3D80%2525%26leftWidth%3D20%2525%26rightWidth%3D0%2525%26showFooter%3Dfalse%26showHeader%3Dfalse%26_adf.ctrl-state%3Dlhbcl1mxr_4

Khong co HSBC, Cake?
```
STT, Mã BIN , Code       , Tên viết tắt      , SWIFT Code, Tên Tổ chức phát hành thẻ
 1 , 970400 , SGICB      , SaigonBank        , SBITVNVX  , TMCP Sài Gòn Công thương
 2 , 970403 , STP        , Sacombank         , SGTTVNVX  , TMCP Sài Gòn Thương tín
 3 , 970405 , VBA        , Agribank          , VBAAVNVX  , Nông nghiệp và Phát triển Nông thôn Việt Nam
 4 , 970406 , DOB        , DongABank         , EACBVNVX  , TMCP Đông Á
 5 , 970407 , TCB        , Techcombank       , VTCBVNVX  , TMCP Kỹ thương
 6 , 970408 , GPB        , GPBank            , GBNKVNVX  , Thương mại TNHH MTV Dầu Khí Toàn Cầu
 7 , 970409 , BAB        , BacABank          , NASCVNVX  , TMCP Bắc Á
 8 , 970410 , SCVN       , StandardChartered , SCBLVNVX  , TNHH MTV Standard Chartered
 9 , 970412 , PVCB       , PVcomBank         , WBVNVNVX  , TMCP Đại chúng Việt Nam
10 , 970414 , Oceanbank  , Oceanbank         , OCBKUS3M  , TNHH MTV Đại Dương
11 , 970415 , ICB        , VietinBank        , ICBVVNVX  , TMCP Công thương Việt Nam
12 , 970416 , ACB        , ACB               , ASCBVNVX  , TMCP Á Châu
13 , 970418 , BIDV       , BIDV              , BIDVVNVX  , Đầu tư và Phát triển Việt Nam
14 , 970419 , NCB        , NCB               , NVBAVNVX  , TMCP Quốc dân
15 , 970421 , VRB        , VRBank            , VRBAVNVX  , liên doanh Việt Nga
16 , 970422 , MB         , MBBank            , MSCBVNVX  , TMCP Quân Đội
17 , 970423 , TPB        , TPBank            , TPBVVNVX  , TMCP Tiên Phong
18 , 970424 , SHBVN      , ShinhanBank       , SHBKVNVX  , TNHH MTV Shinhan Việt Nam
19 , 970425 , ABB        , ABBank            , ABBKVNVX  , TMCP An Bình
20 , 970426 , MSB        , MSB               , MCOBVNVX  , TMCP Hàng Hải
21 , 970427 , VAB        , VietABank         , VNACVNVX  , TMCP Việt Á
22 , 970428 , NAB        , NamABank          , NAMAVNVX  , TMCP Nam Á
23 , 970429 , SCB        , SCB               , SACLVNVX  , TMCP Sài Gòn
24 , 970430 , PGB        , PGBank            , PGBLVNVX  , TMCP Xăng dầu Petrolimex
25 , 970431 , EIB        , Eximbank          , EBVIVNVX  , TMCP Xuất Nhập khẩu Việt Nam
26 , 970432 , VPB        , VPBank            , VPBKVNVX  , TMCP Việt Nam Thịnh Vượng
27 , 970433 , VIETBANK   , VietBank          , VNTTVNVX  , TMCP Việt Nam Thương Tín
28 , 970434 , IVB        , IndovinaBank      , IABBVNVX  , TNHH Indovina
29 , 970436 , VCB        , Vietcombank       , BFTVVNVX  , TMCP Ngoại thương Việt Nam
30 , 970437 , HDB        , HDBank            , HDBCVNVX  , TMCP Phát triển TP.HCM
31 , 970438 , BVB        , BaoVietBank       , BVBVVNVX  , TMCP Bảo Việt
32 , 970439 , PBVN       , PublicBank        , VIDPVNVX  , liên doanh VID PUBLIC BANK
33 , 970440 , SEAB       , SeABank           , SEAVVNVX  , TMCP Đông Nam Á
34 , 970441 , VIB        , VIB               , VNIBVNVX  , TMCP Quốc Tế Việt Nam
35 , 970442 , HLBVN      , HongLeong         , HLBBVNVX  , TNHH MTV Hong Leong Việt Nam
36 , 970443 , SHB        , SHB               , SHBAVNVX  , TMCP Sài Gòn – Hà Nội
37 , 970444 , CBB        , CBBank            , GTBAVNVX  , Thương mại TNHH MTV Xây dựng Việt Nam
38 , 970446 , COOPBANK   , COOPBANK          ,           , Hợp tác xã Việt Nam
39 , 970448 , OCB        , OCB               , ORCOVNVX  , TMCP Phương Đông
40 , 970449 , LPB        , LPBank            , LVBKVNVX  , TMCP Bưu điện Liên Việt (Ngân hàng TMCP Lộc Phát Việt Nam)
41 , 970452 , KLB        , KienLongBank      , KLBKVNVX  , TMCP Kiên Long
42 , 970454 , VCCB       , VietCapitalBank   , VCBCVNVX  , TMCP Bản Việt
43 , 970455 , IBKHN      , IBKHN             , IBKOVNVX  , Công nghiệp Hàn Quốc - Chi nhánh Hà Nội
44 , 970456 , IBKHCM     , IBKHCM            , IBKOVNVX  , Industrial Bank of Korea - Chi nhánh Hồ Chí Minh
45 , 970457 , WVN        , Woori             , HVBKVNVX  , Ngân hàng TNHH Một Thành Viên Woori Bank Việt Nam
46 , 970458 , UOB        , UnitedOverseas    , UOVBVNVX  , Ngân hàng TNHH Một Thành Viên UOB Việt Nam
47 , 970459 , CIMB       , CIMBBank          , CIBBVNVN  , Ngân hàng TNHH Một Thành Viên CIMB Việt Nam
48 , 970460 , Vietcredit , Vietcredit        ,           , Công ty Tài chính cổ phần Xi Măng
49 , 970462 , KBHN       , KookminHN         , CZNBVN2X  , Ngân hàng Kookmin - Chi nhánh Hà Nội
50 , 970463 , KBHCM      , KookminHCM        , CZNBVNVX  , Ngân hàng Kookmin - Chi nhánh Tp. Hồ Chí Minh
51 , 970464 , FCCOM      , TNEXFinance       ,           , Công ty Tài chính TNHH MTV CỘNG ĐỒNG (TNHH MTV TNEX)
52 , 970465 , SINOPAC    , SINOPAC           , SINOVNVX  , Ngân hàng SINOPAC - Chi nhánh Tp. Hồ Chí Minh
53 , 970466 , KEBHANAHCM , KEBHanaHCM        , KOEXVN2X  , Ngân hàng KEB HANA - Chi nhánh Tp. Hồ Chí Minh
54 , 970467 , KEBHANAHN  , KEBHANAHN         , KOEXVNVX  , Ngân hàng KEB HANA - Chi nhánh Hà Nội
55 , 970468 , MAFC       , MAFC              ,           , Công ty Tài chính TNHH MTV Mirae Asset (Việt Nam)
56 , 970470 , MCredit    , MCredit           ,           , Công ty Tài chính TNHH MB SHINSEI
```

### Website của VietQR

Tôi chỉ thấy Napas đề cập tới webiste VietQR và trang web của nó là `vietqr.net`. Nhưng khi search `vietqr` trên Google các kết quả top đầu thường là:
```
vietqr.vn
vietqr.io
vietqr.co
```

Tôi không biết các đơn vị này có thuộc Napas hay không, tôi đã thử tìm kiếm kỹ và không tìm được link nào của Napas đề cập tới những website trên cả. Tôi cho rằng 99% đây là những đơn vị cá nhân độc lập. Tôi viết để cảnh báo bạn hãy thận trọng khi đọc những thông tin từ họ, đừng nhầm tưởng họ là đại diện của Napas.

## Phát triển
### Duy trì bank.csv
File bank.csv chứa thông tin của toàn bộ ngân hàng được cấp mã BIN ở Việt Nam (theo website của Ngân Hàng Nhà Nước Việt Nam), khi file này được cập nhật cần chạy lệnh sau để generate lại file `bank_generated.go`

```sh
$ go run ./cmd/bank.go
```
Khi thay đổi `./bank.csv`

## License [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
MIT
