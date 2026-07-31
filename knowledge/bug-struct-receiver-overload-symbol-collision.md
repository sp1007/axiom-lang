---
name: bug-struct-receiver-overload-symbol-collision
description: overload cùng tên trên receiver struct dùng chung một symbol nên mọi lời gọi bind vào body khai báo trước; -O1 chỉ CHE bằng inlining
metadata:
  type: project
---

# BUG MỞ (hole C) — overload cùng tên trên receiver **struct** đè nhau ở symbol

**Trạng thái:** OPEN (phát hiện 2026-07-31, probe8). Gate khi sửa: **B==C** (chạm tên symbol phát ra).

## Cơ chế
`typecheck.ax:1233-1246` `free_fn_bare_mangles` trả **false** khi param 0 là struct/sum, dựa trên giả
định ghi ngay trong comment: `x86_regs.ax:338` mangle thành `ax_<Struct>_<fn>`, "duy nhất theo
receiver". Giả định đó **đúng theo (receiver, TÊN), không theo CHỮ KÝ**. Hệ quả dây chuyền: vòng
uniquing Phase-3.5 (`typecheck.ax:3227-3240`) không bao giờ gắn `MODDUP` cho overload thứ hai ⇒ hai
hàm phát ra **cùng một symbol** ⇒ **mọi lời gọi bind vào body được khai báo TRƯỚC**.

## ⚠️ Vì sao bug này DỄ BỊ ĐỌC NHẦM (đọc kỹ trước khi đo)
| Probe | Kết quả | Bẫy |
|---|---|---|
| `g10_seq` (`f(s,41) + f(s)`) | **2** ở `-O0`/default · **42** ở `-O1` | **-O1 chỉ CHE bằng inlining.** Ai chỉ đo `-O1` sẽ kết luận "không có bug". |
| `g11_revboth` (đảo thứ tự khai báo) | **82** | cả hai site lấy body 2 tham số ⇒ chứng minh "body khai báo trước thắng", không phải ngẫu nhiên |
| `g5_only2` | 22 — nhưng **11 dưới `-no-dfe`** | các ca "đúng" là **ẢO GIÁC do dead-function elimination** xoá body bị che. DFE che bug. |
| `g1_arity_overload`, `g2_freefn_arity` | 2 | với cả method inline trong struct |
| `g12_intseq`, `g13_strparam` (receiver i64/str) | **42** ✅ | những kiểu này CÓ bare-mangle nên Phase 3.5 unique được ⇒ đúng, và đó là bằng chứng chỉ đúng vào nhánh struct |

**Không phải flake:** deterministic 5/5 lần chạy, dựng lại hai lần, binary **byte-identical**.
Và `default ≡ -O0` vì `optimize` mặc định **false** (`main_air.ax:854/943`) — nên "default" KHÔNG
phải một mức đo độc lập với `-O0`; muốn đối chứng thật thì phải chạy `-O1`.

## Hướng sửa tối thiểu
Bỏ việc coi param 0 kiểu struct/sum là "tự duy nhất": chạy vòng quét va chạm Phase-3.5 trên **mọi**
SYM_FUNC không-extern, không-generic, khoá theo **(name_id, dạng mangle)** — chứ không chỉ trên các
hàm bare-mangle. Chạm typecheck **và** tên symbol phát ra ⇒ **B==C bắt buộc trước commit**, không
được coi là frontend-only.

## Oracle bắt buộc
- Cả `-O0` **và** `-O1` (chỉ `-O0`/default là không đủ để thấy bug biến mất; chỉ `-O1` là không thấy
  bug). Thêm một lượt `-no-dfe` cho ca `g5` — nếu không, DFE sẽ che.
- Thứ tự khai báo **cả xuôi lẫn ngược** (`g10` và `g11`) — chính thứ tự là thứ quyết định kết quả.
- Control chống over-reach: receiver i64/str (`g12`, `g13`), và hai struct khác nhau cùng tên method.

## Liên quan
[[BACKLOG]] (hole C) · `bin/probe8/` (runner `run8.sh`, ma trận `matrix8.sh`) ·
[[bug-mono-generic-ret-typaram-f64]] · [[bug-iface-conformance-unchecked-sites]].
