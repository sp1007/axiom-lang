---
name: lesson-taskstop-leaves-suite-running
description: TaskStop on a background regression run does NOT kill the bash script; it keeps building into the shared REGTMP and corrupts a later run's results
metadata:
  type: feedback
---

# `TaskStop` không giết script — nó vẫn chạy và phá lượt sau (đo 2026-09-05)

**Triệu chứng:** lượt quét `AXEXTRA=-O0` báo **710/711, `FAIL t_hoftup (exit: got '42' want '44')`**.
Chạy lại `t_hoftup` một mình trên **cả** driver trước-fix lẫn sau-fix, ở `-O0`, `-O1`, và đúng tổ hợp
cờ mà suite dùng (`-O1 -O0`) ⇒ **44 cả 6 lần**. Không tái hiện được.

**Không được kết luận "flake"** (CLAUDE.md §24 mục 4: một lần fail là *bug report* cho tới khi quy
được về nguyên nhân có tên). Bằng chứng quyết định nằm ở **chính file exe mà suite đã chạy**, vẫn
còn trên đĩa:

```
/tmp/reg_t_hoftup.exe   25600 byte, 17:38   -> chạy lại NGAY BÂY GIỜ = 44
```

`25600` là kích thước bản dựng **`-O1`**. Lượt `-O0` chỉ sinh ra **26112**. Vậy cái ghi đè lúc 17:38
**không phải** lượt `-O0` đang chạy — mà là một lượt `-O1` khác.

**Nguyên nhân có tên:** trước đó tôi khởi chạy nhầm một lượt suite (dùng biến sai, xem dưới) rồi gọi
`TaskStop`. `TaskStop` kết thúc *task* nhưng **script bash + tiến trình con vẫn sống**. Nó tiếp tục
dựng exe `-O1` vào **cùng đường dẫn dùng chung** `$REGTMP/reg_<name>.exe` mà lượt `-O0` đang đọc.
`t_hoftup` bị **thực thi giữa lúc file đang bị ghi đè** ⇒ đọc ra một giá trị thứ ba (42).
⚠️ `Get-Process` lọc `axc_` **KHÔNG chứng minh được là đã sạch**: giữa hai lần build không có tiến
trình compiler nào tồn tại, nên mẫu tức thời rất dễ trượt.

## Luật rút ra
1. ⛔ **Sau `TaskStop` một lượt suite, phải coi như nó VẪN ĐANG CHẠY.** Kiểm bằng *dấu vết ghi file*
   (mtime/kích thước trong `REGTMP`), không bằng một lần `Get-Process`.
2. ⭐ **Luôn đặt `REGTMP` riêng cho mỗi lượt chạy** (`REGTMP=/tmp/regO0_iso`). Đường dẫn mặc định
   `/tmp` là **dùng chung**, nên hai lượt bất kỳ sẽ giẫm lên nhau — đây chính là hạng mục 3b trong
   CLAUDE.md §24 (exe bị ghi đè trong lúc tiến trình khác nạp nó), nhưng qua một lối vào mới:
   **không cần build song song có chủ ý, chỉ cần một lượt cũ chưa chết.**
3. **Vứt lượt bị giết, đừng union với lượt sạch** — và cũng đừng tin lượt *kế tiếp* nếu nó dùng chung
   `REGTMP` với lượt vừa bị giết.
4. Forensics đúng chỗ: **exe mà suite đã chạy vẫn còn trên đĩa**. So `ls -l` (kích thước) + `cmp` với
   bản dựng sạch trả lời ngay "suite đã chạy đúng file nó vừa dựng hay không" — rẻ hơn mọi suy đoán.

## Lỗi đi kèm (là thứ đã sinh ra lượt chạy thừa)
`scripts/regression_repros.sh` **không đọc `AXFLAGS`**. Biến đúng là **`AXEXTRA`**
(khai báo ở `:23`, nối vào mọi dòng build). Chạy `AXFLAGS=-O0 ...` ⇒ script chạy **y hệt mức mặc
định** và báo GREEN — một lượt "-O0" chưa từng kiểm `-O0`. Cùng hình dạng với sự cố assert cũ trong
handoff: **con số xanh cho một thứ không hề được chạy**.
Ghi nhớ dạng lệnh đúng, có cô lập:
```sh
rm -rf /tmp/regO0_iso && mkdir -p /tmp/regO0_iso
AXC=bin/axc_fpA.exe AXEXTRA=-O0 REGTMP=/tmp/regO0_iso bash scripts/regression_repros.sh
```
⚠️ Lưu ý dòng chính build ở `:1057` là `-O1 $AXEXTRA`, nên `AXEXTRA=-O0` thành **`-O1 -O0`**
(cờ sau thắng — đã đo: 26112 byte, khác hẳn bản `-O1` 25600 byte).

Liên quan: [[lesson-exit-code-8bit-masking]]
