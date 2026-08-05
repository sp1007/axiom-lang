# Session handoff — 2026-07-31a

Trạng thái chuẩn để tiếp tục. Đọc file này + [BACKLOG](BACKLOG.md); **không** đọc `MEMORY.md` nguyên
file (175 KB ≈ 87k token) — chỉ `Grep` theo tên topic.

## Mốc cây (cập nhật cuối phiên)
- HEAD khi mở phiên: `f6ac69e` → nay **`3370f19`**. Commit code của phiên: **`a538983`** (retire rank 1).
- Driver `bin/axc_native.exe` = **A==B `B10DABE6…`** (2.297.856 byte, đã promote + sanity 42);
  mốc B==C gần nhất `D3EABC61` (`b8ac125`).
- **BASELINE = 662/662** (default **và** `-O0`).
- Cây sạch; untracked còn lại: `.claude/settings.json` (của user, đừng đụng).

## Việc đã làm trong phiên
1. **Làm mới BACKLOG** (`bc30f58`) — nó sai ở gần như mọi dòng quan trọng: baseline 611 (thật 649),
   HEAD `5b0eb92c` (thật `f6ac69e`), bug #1 "ĐANG SỬA" (đã ship `6febd02`), bug #3 mở (đã ship
   `e6c507c`), task 0 mở (đã ship `6febd02`). Tất cả đều bị đóng bởi commit **SAU trong CÙNG phiên**
   mà không ai gạch mục backlog — đúng kiểu trôi lệch mà chính header của file cảnh báo.
2. **Bug #4 được ghi lại là GIẢI THỂ (dissolved), không phải "tự nhiên hết"** — xem dưới.

## ⭐ BÀI HỌC QUY TRÌNH — `git log` TRƯỚC KHI dispatch (tôi vi phạm luật đã có sẵn)
Tôi tự đo thấy `r6e.ax` = 42 và `probe4/f1.ax` = 42 (trước là 110), còn compiler trước `6febd02` vẫn
101. Vì **không** commit nào trong khoảng đó nói về thanh ghi trả về của lời gọi gián tiếp, mà
`b8ac125` lại chèn thêm lệnh convert vào luồng float, tôi nghi bug chỉ bị **CHE** do xê dịch cấp
phát thanh ghi (che ≠ sửa) và dispatch investigator đi phán quyết.

**Nghi ngờ đó đúng về nguyên tắc nhưng thừa**: `git log --oneline -14` cho thấy `3741afc` đã **bác
bỏ** cách phát biểu và `a281992` đã **sửa** từ 02:04 cùng ngày. Skill autopilot đã có luật "cross-check
mục backlog với `git log` TRƯỚC khi dispatch" và tôi bỏ qua ⇒ mất một lượt dispatch. Đã `TaskStop`
agent đó và ghi bài học vào BACKLOG.
**Luật rút ra:** ở repo này, backlog trôi lệch trong CÙNG một phiên là trạng thái BÌNH THƯỜNG, không
phải ngoại lệ. Một `git log` rẻ hơn mọi dispatch.

## Sự thật về bug #4 (đừng đi săn lại bug thanh ghi float)
Không phải "f32 return qua dispatch đọc thanh ghi cũ". Phân giải tên method khớp tên trần với
**SUBSTRING chặn bởi `_`** của tên method khác ⇒ **chọn nhầm hàm**: `p32_r32` kết thúc bằng `_r32`
nên **cả hai** slot vtable nhận địa chỉ `p32_r32`, và `i.r32()` gọi hàm 2 tham số qua call site 1 đối
số, trả về một tham số đọc từ thứ lời gọi trước để lại trong XMM1. **Thanh ghi cũ là HỆ QUẢ.**
Chốt bằng disassembly: `mov $0x2a` không xuất hiện trong image ⇒ `S.r32` chưa từng được sinh.
Root cause `ec5667d`, có trước công việc dispatch **hai tháng** ⇒ đó là lý do repro fail y hệt ở CẢ
hai phía của `5359a39`. Bug **type-agnostic**, **không riêng dispatch** (gọi tĩnh cũng sai), dạng
nặng là **nhầm ARITY** và **nhầm KIỂU** (-O0 vs -O1 bất đồng). Refute: `3741afc`; fix: `a281992`.

## ✅ ĐÃ SHIP trong phiên: `a538983` — retire rank 1
Chi tiết đầy đủ trong commit message + [[BACKLOG]]. Ba điều đáng nhớ:
1. **Implementer phản biện 3 điểm trong brief của tôi và ĐÚNG cả 3.** (a) Tôi ghi "a1/d5/d2 phải ra
   42" — sai: không có method `eq`/`lt`/`add` thì **không tồn tại** toán tử, nên **REJECT** mới là
   đúng; 42 chỉ khả dĩ vì cửa sổ lỏng đang trả lời. (b) Bỏ rank 1 làm `+` trên struct thành SIGSEGV;
   agent xác minh segfault đó **có sẵn từ trước**, rồi hoàn thiện nhánh số học RFC 0007 §3.1.
   (c) `h1` là **rank 2**, nên "xoá rank 1" và "h1 phải ra 42" **mâu thuẫn nhau** — agent từ chối mở
   scope, đúng. ⇒ Bài học viết brief: **đừng ghi kỳ vọng số cụ thể cho ca mà semantics đúng là lỗi.**
2. **Tôi tự kiểm chứng độc lập trước khi commit**, không tin báo cáo: md5 A==B, 3 oracle reject không
   sinh exe trong khi compiler CŨ nhận và cho 7/7/8, diagnostic render tận mắt, và breakage audit
   1215 file (7 reject mới đúng dự định, 2 accept mới, 0 collateral).
3. **Bẫy đo của chính tôi**: xem mục "BÀI HỌC ĐO" trong [[BACKLOG]] — union lượt chạy bị giết.

## ✅ ĐÃ SHIP thứ hai: `c3eae77` — hole C (Phase-3.5 khoá theo **nhóm mangle**)
Driver promote sang bản **B==C `52D1ABD4`**, baseline **672/672**. Ba điểm đáng nhớ:
1. **Implementer lại phản biện đúng**: hai shape tôi liệt trong oracle (`g1`, `g2`) là **defect
   KHÁC** và vẫn = 2 — `resolve_method_overload` (`typecheck.ax:1135`) chọn theo **RECEIVER**, không
   hề thấy **arity**, trong khi `resolve_free_call_overload` có gate theo số tham số. Lại đúng hình
   dạng "một luật, hai bản sao, một bản không bao giờ được mở rộng". Quy trách nhiệm **đo được**:
   CÙNG một bản compiler, CÙNG bộ khai báo — dạng gọi tự do ra **42**, dạng method ra **2** ⇒ symbol
   giống hệt, chỉ khác khâu bind. Cần **quyết định semantics** trước ⇒ tách ra
   [[bug-method-call-overload-ignores-arity]], và đặt comment ngay tại site Phase-3.5: *đừng "sửa"
   ở đây*. **Không hàng test nào bị làm yếu** — các hàng đó chưa từng pass.
2. **Oracle phải CALIBRATE, không được pass rỗng**: `t_structoverloaddfe` = **42 trên driver cũ ở
   default** nhưng **1 dưới `-no-dfe`** ⇒ đó là lý do tồn tại khối `-no-dfe` riêng. Tôi tự đo lại.
3. **Tôi chạy bốn suite mà implementer khai là CHƯA chạy** (RFC 0035 §9 gate vào chúng, và thay đổi
   này đụng đúng mangler): `lib_collision` 6/6, `exe_size` 4/4, `ctgc` 16/16, `elf_linux` 12/12.
   ⇒ Bài học: khi agent nói thẳng "cái này tôi chưa chạy", **đó là danh sách việc của mình**, không
   phải câu xã giao.

## 🔄 ĐANG CHẠY khi viết file này
**Investigator — mono `-> T` ↦ f64 trả 0.0** ([[bug-mono-generic-ret-typaram-f64]]). Đã tự xác minh
lại trên driver mới: `i8` = **40** ở CẢ `-O0` lẫn `-O1`; control `i9`/`i5` = 42, `i6`/`i7` = 41.
Brief bắt buộc **phán quyết bằng disassembly** (tiền lệ bug #4: ba lần chẩn đoán "thanh ghi float
cũ" đều sai), và bắt đo **breadth chưa từng đo**: f32, struct, str, tuple, free fn generic.
Lead: họ fix mono đã có (`bf67c7d`, `82d0565`, `ccecc6a`) đều cùng hình dạng "mất width/class của
type argument đã thay" ⇒ nếu lặp lại thì **mở rộng cơ chế đó**, đừng đẻ bản song song.

## 🔄 (lịch sử) investigator probe8 — ĐÃ XONG
**Investigator** (read-only, ghi vào `bin/probe8/`, đã sinh ~60 chương trình probe): câu hỏi là
**fix `a281992` đã phủ MỌI site phân giải tên chưa**, hay còn matcher anh em vẫn khớp theo
substring/prefix/suffix — hình dạng defect tái diễn của repo ("hai bản sao một luật, một bản không
bao giờ được sửa"; 8 bug truy về hình dạng này). Yêu cầu phủ rõ ba dạng nặng: method trả
**con trỏ/struct** ở vị trí che khuất (nguy cơ segfault), lệch **arity**, lệch **kiểu trả về**; và
chiều ngược lại (tightened matcher có thể **over-reject**: `len` và `buf_len` cùng tồn tại, generic
instances, hai struct cùng tên method).

## 🔜 ĐÃ BRIEF SẴN, dispatch ngay khi investigator trả `bin/` (đừng chạy song song!)
**Đóng lỗ hổng bất biến §9 mà `b8ac125` tự khai** — RFC 0006 **§7.3** nói thẳng: *"Chưa có bất biến
§9 cho bề rộng float"*.
- `air_builder.ax:5814` `verify_air_no_int_into_float`: miền trừu tượng `cls[]` chỉ có
  `0=UNKNOWN, 1=INT, 2=FLOAT` (gán ở `:5870-5893`) ⇒ **`OP_COPY` kiểu f64 đọc vreg f32 vẫn LỌT** —
  đúng con bug `b8ac125` vừa vá tay. Tách FLOAT thành F32/F64, bắt `OP_COPY` khi bề rộng target khác
  bề rộng nguồn.
- **Chuẩn nghiệm thu theo đúng tiền lệ của chính RFC**: phải **calibrate bằng cách stub fix thành
  no-op** rồi xem check có báo không (§7.2 đã làm đúng vậy: báo `inst #5 … type_id=10`), **không**
  được chấp nhận "im lặng = đúng". Và phải **im lặng trên cả 1053 hàm** của chính source compiler.
- Guard hiện tại chỉ có `bin/t_f32widen.ax` (đã calibrate: exit 1 trên compiler trước fix).
- Chạm `air_builder.ax` ⇒ RFC ghi gate **B==C**.

## Giới hạn đã biết, đừng đọc im lặng thành bằng chứng
- Site **ĐỐI SỐ** và **FIELD** không được §9 phủ: AIR không mang kiểu tham số trên `OP_CALL`, cũng
  không mang kiểu field trên `OP_SET_FIELD` (type_id ở đó là kiểu STRUCT). `takes_f64(9)` và
  `H(a: 3)` do **oracle** canh. Mở rộng AIR để mang các kiểu đó = thay đổi IR ⇒ **cần RFC riêng**.
- Bảy vị trí f32→f64 mà `b8ac125` làm cho trả **3 thay vì 0** **KHÔNG phải là đã phủ** — chỉ chuyển
  từ *accept-then-miscompile* sang *accept-without-diagnostic*.

## Nhắc vận hành
- Driver cần subcommand `build`: `bin/axc_native.exe build f.ax -o out.exe -O0`. Thiếu `build` ⇒
  exit 0 mà **không sinh file** (tôi đã suýt mắc lại).
- ⛔ **Serialize mọi thứ ghi `bin/` hoặc chạy suite.** Song song = exe bị cắt cụt 25–30 KB đọc như
  "hồi quy thảm khốc", hoặc suite chết giữa chừng. Chỉ song song hoá investigation read-only.
- Monitor heartbeat 5 phút đang chạy (task `b81z4z1jx`) — **không** `TaskStop` trừ khi user bảo dừng.
