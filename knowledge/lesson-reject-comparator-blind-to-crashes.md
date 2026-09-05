---
name: lesson-reject-comparator-blind-to-crashes
description: The regression suite's cmp=reject decides purely on "no exe emitted", so a compiler SIGSEGV PASSES a reject row — 107 rows are blind to the compiler crashing
metadata:
  type: feedback
---

# `cmp=reject` **KHÔNG phân biệt được "từ chối sạch" với "compiler SẬP"** (đo 2026-09-05)

`scripts/regression_repros.sh:1058-1061`:
```sh
if [ "$cmp" = "reject" ]; then
  # Expect the compiler to REJECT this program (diagnostic, no exe emitted).
  if [ ! -f "$out_exe" ]; then echo "PASS $name (rejected)"; ...
```
Phán quyết **chỉ dựa trên "không có file exe"**. Mà **một SIGSEGV cũng không sinh exe**.
⇒ **compiler sập ⇒ hàng `reject` vẫn PASS.**

📏 **Đo được: 107 hàng `reject`** trong suite — **tất cả đều mù** với việc compiler sập.
Chúng cũng mù với: exit 0 kèm "không sinh exe vì lý do khác", và với việc **chẩn đoán biến mất**
(không hàng nào kiểm rằng có in ra thông báo lỗi nào cả).

## Vì sao phát hiện được (và suýt thì không)
Tôi giao cho implementer đề bài B3b-3 kèm gợi ý *"`cmp=reject` là comparator đúng"*. **Gợi ý đó SAI.**
Implementer đã **đo hiệu chuẩn** thay vì tin lời tôi, thấy hàng `reject` **PASS trên driver TRƯỚC khi
sửa** (lúc đó compiler đang 139), và chuyển sang khối kiểm exit-code theo tiền lệ `input-halt`
(`:1348-1379`). ⇒ **Chính quy tắc "hiệu chuẩn bắt buộc: test phải FAIL được trước khi sửa" đã cứu bài
này.** Không có nó, B3b-3 đã ship kèm một test **không thể fail**.

## Vì sao đây là defect nghiêm trọng, không phải chuyện thẩm mỹ
Suite này là **cổng chính** cho mọi thay đổi compiler. 107 hàng trong đó hiện không thể phát hiện
được **regression tệ nhất có thể** — compiler sập trên đúng chương trình mà chúng canh gác. Một thay
đổi làm 107 chương trình đó chuyển từ "reject sạch" sang "SIGSEGV" sẽ báo **GREEN toàn tập**.
Đúng họ [[lesson-exit-code-8bit-masking]]: **dụng cụ đo âm thầm làm khác điều nó có vẻ làm.**

## Sản phẩm bền phải là một CHECK, không phải ghi chú này
Theo đúng tiền lệ **EXIT-CODE RANGE GUARD** (đã có ở đầu vòng lặp row): sửa nhánh `reject` để ngoài
"không có exe" còn đòi hỏi:
1. **exit code KHÔNG phải tín hiệu crash** — `139` (SIGSEGV), `134` (SIGABRT), `136` (SIGFPE),
   và `124` (timeout của `timeout`); và
2. **exit code khác 0** (từ chối phải *thất bại*, không được âm thầm thành công); và
3. (nên có) **có in ra ít nhất một dòng chẩn đoán** — bắt được ca "reject im lặng", vốn vi phạm §8.

⚠️ **Phải hiệu chuẩn chính cái guard đó**: tiêm một hàng `reject` trỏ vào chương trình làm compiler
sập ⇒ phải FAIL; bỏ ra ⇒ suite chạy sạch. (Guard exit-code range đã được hiệu chuẩn đúng kiểu này.)
⚠️ **Phải chạy toàn suite sau khi thêm**: nếu trong 107 hàng đó **đang có** hàng nào thực chất là
crash được che giấu, guard sẽ làm nó đỏ lên — **đó là tính năng, không phải hỏng**, nhưng phải điều
tra từng cái chứ đừng nới guard cho vừa.
⛔ **Sửa `scripts/regression_repros.sh` TRƯỚC khi bắt đầu lượt chạy nào, không bao giờ trong lúc chạy**
(bash đọc script dần — xem [[lesson-taskstop-leaves-suite-running]]).

## ⭐ HIỆU CHUẨN ĐÃ ĐO TRƯỚC (2026-09-05, trên driver `29DB6D68…`) — guard là THUẦN BỔ SUNG

Câu hỏi chặn đường của bản ghi này (*"nếu trong 107 hàng đang có crash bị che thì guard sẽ làm đỏ"*)
**đã được trả lời bằng phép đo, TRƯỚC khi sửa script**: chạy cả **107** chương trình `reject` qua một
**bản sao riêng** của driver (tách khỏi `bin/` để không đụng lần build nào đang chạy —
[[lesson-taskstop-leaves-suite-running]]):

| tiêu chí | kết quả trên 107/107 |
|---|---|
| exit code | **exit=1** ở **tất cả** — không hàng nào là 139/134/136/124 |
| sinh exe | **không hàng nào** sinh exe |
| dòng chẩn đoán | **≥ 2 dòng** ở tất cả (phân bố 2/3/5/7/8/9/13/29) — **không hàng nào im lặng** |

⇒ **Cả ba tiêu chí đều cài được mà KHÔNG làm đỏ hàng nào.** Guard này là **thuần bổ sung bảo vệ**,
không phải một cuộc dọn dẹp ⇒ nó **không còn bị chặn** bởi việc "phải điều tra từng cái trong 107".

### Fixture hiệu chuẩn (bắt buộc: test phải FAIL được trước khi tin nó)
```axiom
import std.process        // ⇒ SIGSEGV 139, không sinh exe  (đo 2026-09-05)
fn main() -> i32:
    return 0
```
Hôm nay một hàng `reject` trỏ vào chương trình này **PASS** — đúng cái defect này.

### ⚠️ ĐÍNH CHÍNH cho chính đề xuất ở trên: tiêu chí 3 **KHÔNG** bắt được ca này
Bản `std.process` **có in chẩn đoán** (`error: undefined name 'raw32'` × nhiều) **rồi mới sập**.
⇒ Tiêu chí *"có ít nhất một dòng chẩn đoán"* sẽ cho nó **QUA**. Chỉ **tiêu chí 1 (mã thoát là tín
hiệu crash)** mới bắt được. Tiêu chí 3 vẫn đáng có (nó canh §8 "reject im lặng"), nhưng **đừng nhầm
nó là cái chặn crash** — đúng lớp sai lầm mà chính bản ghi này cảnh báo: *một comparator đặt tên theo
Ý ĐỊNH nhưng cài đặt theo TÁC DỤNG PHỤ*.

## Bài học phương pháp
⭐ **Một comparator được đặt tên theo Ý ĐỊNH ("reject") nhưng cài đặt theo TÁC DỤNG PHỤ ("không có
file") sẽ nhận nhầm mọi thứ khác cũng gây ra tác dụng phụ đó.** Trước khi tin một comparator, hãy hỏi:
*"còn nguyên nhân nào khác tạo ra đúng dấu hiệu này?"* Ở đây câu trả lời — **crash** — chính là thứ
tệ nhất trong các nguyên nhân.
⭐ Và: **đừng tin gợi ý comparator trong đề bài của chính mình** — tôi đã đưa gợi ý sai; chỉ phép đo
hiệu chuẩn mới bắt được.

Liên quan: [[lesson-exit-code-8bit-masking]] · [[lesson-comment-protects-one-line-only]] ·
[[lesson-taskstop-leaves-suite-running]]
