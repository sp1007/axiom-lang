---
name: bug-negative-literal-compare-o0
description: "OPEN BUG (probe-found 2026-07-30) — so sánh với LITERAL ÂM có |v| > 2^31 cho kết quả SAI ở -O0 (đúng ở -O1/-O2/-O3). Root: unary minus rồi wrap-to-width về i32 (neg; shl 32; sar 32). Silent miscompile."
metadata:
  type: project
---

# 🐞 OPEN — so sánh với literal ÂM ngoài dải i32 SAI ở `-O0`

**Tình trạng: OPEN, chưa sửa.** Tìm được bằng probe 2026-07-30 (driver `0E24570B`), khi kiểm
tra chéo các thay đổi codegen của phiên. **KHÔNG do các thay đổi phiên này gây ra** — chúng chỉ
động tới peephole ALU/IMUL/copy và căn lề, không động tới kiểu của literal hay `emit_wrap_to_width`.

## Reproducer tối giản
```axiom
fn main() -> i64:
    let c: i64 = -3000000000
    if c == -3000000000:
        return 42
    return 11
```
`-O0` → **11 (SAI)**; `-O1/-O2/-O3` → 42 (đúng).

## Phạm vi — đã khoanh bằng thí nghiệm, không phải suy đoán
| ca | kết quả |
|---|---|
| `c == -5737418117` (âm, \|v\| > 2^32) | ❌ SAI ở -O0 |
| `c == -3000000000` (âm, 2^31 < \|v\| < 2^32) | ❌ SAI ở -O0 |
| `c == -2147483649` (âm, vừa vượt i32) | ❌ SAI ở -O0 |
| `c == 5737418117` (**DƯƠNG** lớn) | ✅ đúng |
| `c == d` (biến vs biến, cả hai âm lớn) | ✅ đúng |
| `c < -2000000000` (âm nhưng **VỪA** i32) | ✅ đúng |
| `let c: i64 = -5737418117` rồi đọc `c` (**chỉ BINDING**) | ✅ đúng |

⇒ Trigger hẹp: **literal ÂM, |v| ngoài dải i32, đứng TRỰC TIẾP trong biểu thức so sánh, ở -O0.**
Binding thì đúng; chỉ literal ở vị trí toán hạng so sánh mới hỏng.

## Root cause — ĐỌC TỪ DISASSEMBLY, không phải giả thuyết
`-O0` phát cho toán hạng phải của phép so sánh:
```
neg    %rdx
shl    $0x20,%rdx      <- dịch trái 32
sar    $0x20,%rdx      <- dịch phải SỐ HỌC 32   ==> SIGN-EXTEND TỪ 32 BIT
cmp    %rdx,%rax
```
`shl 32 ; sar 32` chính là **wrap về i32 rồi sign-extend** — đúng thành ngữ của
`emit_wrap_to_width` → `emit_load_extend` (RFC 0006 / BUG#36, `x86_selector.ax` ~L887).
Nghĩa là: literal dương `3000000000` bị **unary minus** rồi kết quả bị **wrap theo bề rộng i32**
⇒ mọi bit trên 32 bị xoá.
`-O1+` **KHÔNG** phát chuỗi này — nó materialise thẳng `movabs $0xfffffffeaa05f27b` ⇒
**pass const-fold ở tầng SSA che mất bug**, đó là lý do nó chỉ lộ ở -O0.

## Hướng sửa (chưa làm — cần phiên mới, đụng typecheck)
Nghi vấn: literal trong `-<literal>` được suy kiểu **i32 mặc định** (int-literal default), rồi
`emit_wrap_to_width` áp đúng bề rộng đã suy đó. Ở nhánh binding thì kiểu đích `i64` ép được,
nhưng ở **nhánh so sánh** thì hint kiểu không truyền xuống — **cùng họ với
[[bug-f32-compare-float-literal]]**: quy tắc coercion có ở nhánh arithmetic nhưng **thiếu ở
nhánh comparison**. Bài học đã bank ở đó: *"rule coercion thêm ở 1 nhánh binop mà quên nhánh
sibling = partial fix; áp cho MỌI nhánh"* — lần này là **cùng lỗi, với int literal âm**.
⇒ Chỗ cần xem trước: nhánh `op == 1` (comparison) trong typecheck, nơi f32 từng phải vá y hệt.

⚠️ **Đây là SILENT MISCOMPILE** (§7 correctness > tất cả), nhưng chỉ ở `-O0`, và mọi gate hiện
tại chạy ở `-O1` trở lên nên 564/564 **không hề thấy**. ⇒ Khi sửa, **oracle phải chạy CẢ -O0**.
⭐ Bài học hạ tầng: **suite regression phần lớn không chạy -O0 ⇒ cả một lớp bug chỉ-ở-O0 không
có coverage.** Đáng cân nhắc thêm vài dòng `-O0` vào `regression_repros.sh`.
