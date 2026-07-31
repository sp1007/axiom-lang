---
name: bug-struct-receiver-overload-symbol-collision
description: overload cùng tên trên receiver struct dùng chung một symbol nên mọi lời gọi bind vào body khai báo trước; -O1 chỉ CHE bằng inlining
metadata:
  type: project
---

# BUG (hole C) — overload cùng tên trên receiver **struct** đè nhau ở symbol

**Trạng thái:** ✅ **ĐÃ SỬA 2026-07-31** ở dạng gọi **TỰ DO** — `fn_mangle_group` thay
`free_fn_bare_mangles`, vòng Phase-3.5 nay khoá theo **(name_id, nhóm mangle)**. Gate thực đo:
**A==B==C `52D1ABD4AE9E6EF11216AD3B8318D1592C1C03F383D49F5464B6ABF0A6C9478B`** (inert với chính
compiler ⇒ bằng chứng là ORACLE chứ không phải gate, đúng cảnh báo RFC 0035 §7bis).
Oracle: `bin/t_structoverload.ax` + `bin/t_structoverloaddfe.ax` (1 → 42). Spec: RFC 0035 §2bis.

⚠️ **CÒN MỞ, và KHÔNG phải bug này:** lời gọi viết bằng **cú pháp method** (`s.f(41)`, method
inline, dạng tĩnh `S.f(&s,41)`) vẫn bind sai vì `resolve_method_overload` chọn theo RECEIVER,
bỏ qua arity ⇒ [[bug-method-call-overload-ignores-arity]]. Quy trách nhiệm chắc chắn: cùng bộ
khai báo, **một bản compiler**, dạng tự do ra 42 còn dạng method ra 2 ⇒ symbol giống hệt nhau,
chỉ khác khâu bind.

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

## Cách đã sửa (2026-07-31)
Bỏ việc coi param 0 kiểu struct/sum là "tự duy nhất". `free_fn_bare_mangles` (trả bool) trở thành
`fn_mangle_group` (trả u32): `0` = không tham gia (extern / đã MODDUP / generic template / tên đặc
biệt), `1` = `ax_<name>`, `1 + <name_id của kiểu receiver>` = `ax_<Struct>_<fn>`. Vòng Phase-3.5
gắn MODDUP khi trùng **CẢ** name_id **LẪN** nhóm ⇒ hai receiver KHÁC nhau nằm khác nhóm nên method
trùng tên trên hai struct không bị đụng tới. Chạm tên symbol phát ra ⇒ **B==C bắt buộc** (đã đo,
xem trên). Nhánh `unit_qualifier` cố tình KHÔNG mô hình hoá: nó được kiểm TRƯỚC MODDUP ở
`x86_regs.ax:277` nên tag bị bỏ qua với những symbol đó.

## Oracle bắt buộc
- Cả `-O0` **và** `-O1` (chỉ `-O0`/default là không đủ để thấy bug biến mất; chỉ `-O1` là không thấy
  bug). Thêm một lượt `-no-dfe` cho ca `g5` — nếu không, DFE sẽ che.
- Thứ tự khai báo **cả xuôi lẫn ngược** (`g10` và `g11`) — chính thứ tự là thứ quyết định kết quả.
- Control chống over-reach: receiver i64/str (`g12`, `g13`), và hai struct khác nhau cùng tên method.

## Liên quan
[[BACKLOG]] (hole C) · `bin/probe8/` (runner `run8.sh`, ma trận `matrix8.sh`) ·
[[bug-mono-generic-ret-typaram-f64]] · [[bug-iface-conformance-unchecked-sites]].
