#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_PathBuf;
struct ax_FileMetadata;
struct ax__AX_std_Result__FileMetadata__string;
struct ax__AX_std_Option__PathBuf;
struct ax__AX_std_Result__void__string;
struct ax__AX_std_Option__string;
struct ax__AX_std_HashMap__string__string;
struct ax_Command;
struct ax_Child;
struct ax__AX_std_Result__Child__string;
struct ax_ExitStatus;
struct ax__AX_std_Result__ExitStatus__string;
struct ax_Output;
struct ax__AX_std_Result__Output__string;
struct ax__AX_std_Result__u64__string;

/* Type definitions */
struct ax_PathBuf {
    ax_string inner;
};

struct ax_FileMetadata {
    ax_u64 size;
    ax_u64 modified;
    ax_u64 created;
    ax_bool is_file;
    ax_bool is_dir;
    ax_bool is_symlink;
    ax_u32 mode;
};

enum ax__AX_std_Result__FileMetadata__string_tag {
    ax__AX_std_Result__FileMetadata__string_Ok = 0,
    ax__AX_std_Result__FileMetadata__string_Err = 1,
};

struct ax__AX_std_Result__FileMetadata__string {
    enum ax__AX_std_Result__FileMetadata__string_tag tag;
    union {
        struct ax_FileMetadata Ok;
        ax_string Err;
    } data;
};

static inline struct ax__AX_std_Result__FileMetadata__string ax__AX_std_Result__FileMetadata__string_ok(struct ax_FileMetadata value) {
    struct ax__AX_std_Result__FileMetadata__string _result;
    _result.tag = ax__AX_std_Result__FileMetadata__string_Ok;
    _result.data.Ok = value;
    return _result;
}

static inline struct ax__AX_std_Result__FileMetadata__string ax__AX_std_Result__FileMetadata__string_err(ax_string value) {
    struct ax__AX_std_Result__FileMetadata__string _result;
    _result.tag = ax__AX_std_Result__FileMetadata__string_Err;
    _result.data.Err = value;
    return _result;
}

enum ax__AX_std_Option__PathBuf_tag {
    ax__AX_std_Option__PathBuf_Some = 0,
    ax__AX_std_Option__PathBuf_None = 1,
};

struct ax__AX_std_Option__PathBuf {
    enum ax__AX_std_Option__PathBuf_tag tag;
    union {
        struct ax_PathBuf Some;
    } data;
};

static inline struct ax__AX_std_Option__PathBuf ax__AX_std_Option__PathBuf_some(struct ax_PathBuf value) {
    struct ax__AX_std_Option__PathBuf _result;
    _result.tag = ax__AX_std_Option__PathBuf_Some;
    _result.data.Some = value;
    return _result;
}

static inline struct ax__AX_std_Option__PathBuf ax__AX_std_Option__PathBuf_none(void) {
    struct ax__AX_std_Option__PathBuf _result;
    _result.tag = ax__AX_std_Option__PathBuf_None;
    return _result;
}

enum ax__AX_std_Result__void__string_tag {
    ax__AX_std_Result__void__string_Ok = 0,
    ax__AX_std_Result__void__string_Err = 1,
};

struct ax__AX_std_Result__void__string {
    enum ax__AX_std_Result__void__string_tag tag;
    union {
        ax_string Err;
    } data;
};

static inline struct ax__AX_std_Result__void__string ax__AX_std_Result__void__string_ok(void) {
    struct ax__AX_std_Result__void__string _result;
    _result.tag = ax__AX_std_Result__void__string_Ok;
    return _result;
}

static inline struct ax__AX_std_Result__void__string ax__AX_std_Result__void__string_err(ax_string value) {
    struct ax__AX_std_Result__void__string _result;
    _result.tag = ax__AX_std_Result__void__string_Err;
    _result.data.Err = value;
    return _result;
}

enum ax__AX_std_Option__string_tag {
    ax__AX_std_Option__string_Some = 0,
    ax__AX_std_Option__string_None = 1,
};

struct ax__AX_std_Option__string {
    enum ax__AX_std_Option__string_tag tag;
    union {
        ax_string Some;
    } data;
};

static inline struct ax__AX_std_Option__string ax__AX_std_Option__string_some(ax_string value) {
    struct ax__AX_std_Option__string _result;
    _result.tag = ax__AX_std_Option__string_Some;
    _result.data.Some = value;
    return _result;
}

static inline struct ax__AX_std_Option__string ax__AX_std_Option__string_none(void) {
    struct ax__AX_std_Option__string _result;
    _result.tag = ax__AX_std_Option__string_None;
    return _result;
}

struct ax__AX_std_HashMap__string__string {
    ax_string* keys;
    ax_string* values;
    ax_u64* hashes;
    ax_bool* occupied;
    ax_i64 size;
    ax_i64 cap;
};

struct ax_Command {
    ax_string program;
    ax_vec args;
    struct ax__AX_std_HashMap__string__string env;
    struct ax__AX_std_Option__string cwd;
};

struct ax_Child {
    ax_i64 pid;
};

enum ax__AX_std_Result__Child__string_tag {
    ax__AX_std_Result__Child__string_Ok = 0,
    ax__AX_std_Result__Child__string_Err = 1,
};

struct ax__AX_std_Result__Child__string {
    enum ax__AX_std_Result__Child__string_tag tag;
    union {
        struct ax_Child Ok;
        ax_string Err;
    } data;
};

static inline struct ax__AX_std_Result__Child__string ax__AX_std_Result__Child__string_ok(struct ax_Child value) {
    struct ax__AX_std_Result__Child__string _result;
    _result.tag = ax__AX_std_Result__Child__string_Ok;
    _result.data.Ok = value;
    return _result;
}

static inline struct ax__AX_std_Result__Child__string ax__AX_std_Result__Child__string_err(ax_string value) {
    struct ax__AX_std_Result__Child__string _result;
    _result.tag = ax__AX_std_Result__Child__string_Err;
    _result.data.Err = value;
    return _result;
}

struct ax_ExitStatus {
    ax_i32 code;
};

enum ax__AX_std_Result__ExitStatus__string_tag {
    ax__AX_std_Result__ExitStatus__string_Ok = 0,
    ax__AX_std_Result__ExitStatus__string_Err = 1,
};

struct ax__AX_std_Result__ExitStatus__string {
    enum ax__AX_std_Result__ExitStatus__string_tag tag;
    union {
        struct ax_ExitStatus Ok;
        ax_string Err;
    } data;
};

static inline struct ax__AX_std_Result__ExitStatus__string ax__AX_std_Result__ExitStatus__string_ok(struct ax_ExitStatus value) {
    struct ax__AX_std_Result__ExitStatus__string _result;
    _result.tag = ax__AX_std_Result__ExitStatus__string_Ok;
    _result.data.Ok = value;
    return _result;
}

static inline struct ax__AX_std_Result__ExitStatus__string ax__AX_std_Result__ExitStatus__string_err(ax_string value) {
    struct ax__AX_std_Result__ExitStatus__string _result;
    _result.tag = ax__AX_std_Result__ExitStatus__string_Err;
    _result.data.Err = value;
    return _result;
}

struct ax_Output {
    struct ax_ExitStatus status;
    ax_string stdout_str;
    ax_string stderr_str;
};

enum ax__AX_std_Result__Output__string_tag {
    ax__AX_std_Result__Output__string_Ok = 0,
    ax__AX_std_Result__Output__string_Err = 1,
};

struct ax__AX_std_Result__Output__string {
    enum ax__AX_std_Result__Output__string_tag tag;
    union {
        struct ax_Output Ok;
        ax_string Err;
    } data;
};

static inline struct ax__AX_std_Result__Output__string ax__AX_std_Result__Output__string_ok(struct ax_Output value) {
    struct ax__AX_std_Result__Output__string _result;
    _result.tag = ax__AX_std_Result__Output__string_Ok;
    _result.data.Ok = value;
    return _result;
}

static inline struct ax__AX_std_Result__Output__string ax__AX_std_Result__Output__string_err(ax_string value) {
    struct ax__AX_std_Result__Output__string _result;
    _result.tag = ax__AX_std_Result__Output__string_Err;
    _result.data.Err = value;
    return _result;
}

enum ax__AX_std_Result__u64__string_tag {
    ax__AX_std_Result__u64__string_Ok = 0,
    ax__AX_std_Result__u64__string_Err = 1,
};

struct ax__AX_std_Result__u64__string {
    enum ax__AX_std_Result__u64__string_tag tag;
    union {
        ax_u64 Ok;
        ax_string Err;
    } data;
};

static inline struct ax__AX_std_Result__u64__string ax__AX_std_Result__u64__string_ok(ax_u64 value) {
    struct ax__AX_std_Result__u64__string _result;
    _result.tag = ax__AX_std_Result__u64__string_Ok;
    _result.data.Ok = value;
    return _result;
}

static inline struct ax__AX_std_Result__u64__string ax__AX_std_Result__u64__string_err(ax_string value) {
    struct ax__AX_std_Result__u64__string _result;
    _result.tag = ax__AX_std_Result__u64__string_Err;
    _result.data.Err = value;
    return _result;
}


/* Global variables */
extern const ax_bool ax_IS_LINUX;
const ax_bool ax_IS_LINUX = 0;
extern const ax_bool ax_IS_MACOS;
const ax_bool ax_IS_MACOS = 0;
extern const ax_bool ax_IS_WINDOWS;
const ax_bool ax_IS_WINDOWS = 1;
extern const ax_string ax_OS_NAME;
const ax_string ax_OS_NAME = AX_STR("windows");
extern const ax_string ax_ARCH_NAME;
const ax_string ax_ARCH_NAME = AX_STR("x86_64");
extern const ax_string ax_PATH_SEP;
const ax_string ax_PATH_SEP = AX_STR("\\");
extern const ax_i32 ax_SIGINT;
const ax_i32 ax_SIGINT = 2;
extern const ax_i32 ax_SIGTERM;
const ax_i32 ax_SIGTERM = 15;
extern const ax_i32 ax_SIGHUP;
const ax_i32 ax_SIGHUP = 1;
extern const ax_i32 ax_SIGUSR1;
const ax_i32 ax_SIGUSR1 = 10;
extern const ax_i32 ax_SIGUSR2;
const ax_i32 ax_SIGUSR2 = 12;

/* Function prototypes */
ax_bool ax_sum_layout_is_pointer(void);
struct ax_PathBuf ax_pathbuf_new(ax_string s);
struct ax_PathBuf ax_PathBuf_pathbuf_join(struct ax_PathBuf self, ax_string component);
struct ax__AX_std_Option__PathBuf ax_PathBuf_pathbuf_parent(struct ax_PathBuf self);
struct ax__AX_std_Option__string ax_PathBuf_pathbuf_file_name(struct ax_PathBuf self);
struct ax__AX_std_Option__string ax_PathBuf_pathbuf_extension(struct ax_PathBuf self);
struct ax_PathBuf ax_PathBuf_pathbuf_with_extension(struct ax_PathBuf self, ax_string ext);
ax_bool ax_PathBuf_pathbuf_is_absolute(struct ax_PathBuf self);
ax_string ax_PathBuf_pathbuf_to_str(struct ax_PathBuf self);
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
static ax_u8* ax_os_str_to_c_str(ax_string s);
struct ax__AX_std_Result__FileMetadata__string ax_metadata(ax_string path);
ax_bool ax_exists(ax_string path);
struct ax__AX_std_Result__void__string ax_create_dir(ax_string path);
struct ax__AX_std_Result__void__string ax_create_dir_all(ax_string path);
struct ax__AX_std_Result__void__string ax_remove_file(ax_string path);
struct ax__AX_std_Result__void__string ax_remove_dir(ax_string path);
struct ax__AX_std_Result__void__string ax_rename(ax_string from, ax_string to);
struct ax__AX_std_Result__u64__string ax_copy(ax_string from, ax_string to);
struct ax_PathBuf ax_temp_dir(void);
void ax_trap_signal(ax_i32 sig, void (*handler)(void));
ax_vec ax_parse_cmdline(ax_u8* cmdline);
ax_vec ax_parse_cmdline_w(ax_u16* cmdline);
ax_vec ax_parse_proc_cmdline(ax_u8* buf, ax_i64 len);
static ax_i64 ax_linux_read_file(ax_string path, ax_u8* buf, ax_i64 max_len);
ax_vec ax_args(void);
struct ax__AX_std_Option__string ax_env(ax_string key);
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
struct ax_Command ax_new_command(ax_string program);
void ax_Command_arg(struct ax_Command* self, ax_string arg);
void ax_Command_env(struct ax_Command* self, ax_string key, ax_string value);
void ax_Command_cwd(struct ax_Command* self, ax_string dir);
struct ax__AX_std_Result__Child__string ax_Command_spawn_process(struct ax_Command self);
struct ax__AX_std_Result__Output__string ax_Command_output(struct ax_Command self);
struct ax__AX_std_Result__ExitStatus__string ax_Child_wait(struct ax_Child* self);
struct ax__AX_std_Result__void__string ax_Child_kill(struct ax_Child self);
ax_bool ax_ExitStatus_success(struct ax_ExitStatus self);
ax_i32 ax_ExitStatus_code(struct ax_ExitStatus self);
void exit(ax_i32 code);
void abort(void);
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
static void ax_test_print_str(ax_string s);
static void ax_test_process_spawn(void);
ax_i32 ax_main_usr(void);
ax_bool ax_AX_std_is_ok__FileMetadata__string(struct ax__AX_std_Result__FileMetadata__string self);
ax_bool ax_AX_std_is_some__PathBuf(struct ax__AX_std_Option__PathBuf self);
struct ax_PathBuf ax_AX_std_unwrap__PathBuf(struct ax__AX_std_Option__PathBuf self);
ax_bool ax_AX_std_is_err__void__string(struct ax__AX_std_Result__void__string self);
ax_bool ax_AX_std_is_some__string(struct ax__AX_std_Option__string self);
ax_string ax_AX_std_unwrap__string(struct ax__AX_std_Option__string self);
void ax__AX_std_Vec__u8_push(ax_vec* self, ax_u8 item);
void ax__AX_std_Vec__string_push(ax_vec* self, ax_string item);
void ax__AX_std_Vec__u8_destroy(ax_vec* self);
void ax__AX_std_Vec__u16_push(ax_vec* self, ax_u16 item);
void ax__AX_std_Vec__u16_destroy(ax_vec* self);
ax_vec ax_AX_std_new_vec__string(void);
struct ax__AX_std_HashMap__string__string ax_AX_std_new_hashmap__string__string(void);
void ax__AX_std_HashMap__string__string_insert(struct ax__AX_std_HashMap__string__string* self, ax_string key, ax_string value);
static ax_u64 ax_AX_std_hash_key__string(ax_string key);
struct ax__AX_std_Option__string ax__AX_std_Vec__string_get(ax_vec self, ax_i64 index);
ax_i64 ax_std_string_len(ax_string p0);
ax_string ax_std_string_slice(ax_string p0, ax_i64 p1, ax_i64 p2);
ax_string ax_std_string_concat(ax_string p0, ax_string p1);
ax_bool ax_std_string_starts_with(ax_string p0, ax_string p1);


struct ax_PathBuf ax_pathbuf_new(ax_string s) {
    return ((struct ax_PathBuf){.inner=s});
}

struct ax_PathBuf ax_PathBuf_pathbuf_join(struct ax_PathBuf self, ax_string component) {
    return ((struct ax_PathBuf){.inner=ax_str_concat(ax_str_concat(self.inner, ax_PATH_SEP), component)});
}

struct ax__AX_std_Option__PathBuf ax_PathBuf_pathbuf_parent(struct ax_PathBuf self) {
    ax_string s = self.inner;
    ax_i64 last_sep = ((ax_i64)((-1)));
    ax_i64 i = ((ax_i64)(0));
    while ((i < s.len)) {
        ax_u8* ch = ((ax_u8*)((((ax_i64)(s.ptr)) + i)));
        if ((((*((ax_u8*)(ch))) == ((ax_u8)('/'))) || ((*((ax_u8*)(ch))) == ((ax_u8)('\\'))))) {
            last_sep = i;
        }
        i = (i + 1);
    }
    if ((last_sep >= 0)) {
        return ax__AX_std_Option__PathBuf_some(ax_pathbuf_new(ax_str_slice(s, 0, last_sep)));
    }
    return ax__AX_std_Option__PathBuf_none();
}

struct ax__AX_std_Option__string ax_PathBuf_pathbuf_file_name(struct ax_PathBuf self) {
    ax_string s = self.inner;
    ax_i64 last_sep = ((ax_i64)((-1)));
    ax_i64 i = ((ax_i64)(0));
    while ((i < s.len)) {
        ax_u8* ch = ((ax_u8*)((((ax_i64)(s.ptr)) + i)));
        if ((((*((ax_u8*)(ch))) == ((ax_u8)('/'))) || ((*((ax_u8*)(ch))) == ((ax_u8)('\\'))))) {
            last_sep = i;
        }
        i = (i + 1);
    }
    if ((last_sep >= 0)) {
        return ax__AX_std_Option__string_some(ax_str_slice(s, (last_sep + 1), s.len));
    }
    return ax__AX_std_Option__string_some(s);
}

struct ax__AX_std_Option__string ax_PathBuf_pathbuf_extension(struct ax_PathBuf self) {
    ax_string s = self.inner;
    ax_i64 last_sep = ((ax_i64)((-1)));
    ax_i64 last_dot = ((ax_i64)((-1)));
    ax_i64 i = ((ax_i64)(0));
    while ((i < s.len)) {
        ax_u8* ch = ((ax_u8*)((((ax_i64)(s.ptr)) + i)));
        if ((((*((ax_u8*)(ch))) == ((ax_u8)('/'))) || ((*((ax_u8*)(ch))) == ((ax_u8)('\\'))))) {
            last_sep = i;
        } else if (((*((ax_u8*)(ch))) == ((ax_u8)('.')))) {
            last_dot = i;
        }
        i = (i + 1);
    }
    if ((last_dot > last_sep)) {
        return ax__AX_std_Option__string_some(ax_str_slice(s, (last_dot + 1), s.len));
    }
    return ax__AX_std_Option__string_none();
}

struct ax_PathBuf ax_PathBuf_pathbuf_with_extension(struct ax_PathBuf self, ax_string ext) {
    ax_string s = self.inner;
    ax_i64 last_sep = ((ax_i64)((-1)));
    ax_i64 last_dot = ((ax_i64)((-1)));
    ax_i64 i = ((ax_i64)(0));
    while ((i < s.len)) {
        ax_u8* ch = ((ax_u8*)((((ax_i64)(s.ptr)) + i)));
        if ((((*((ax_u8*)(ch))) == ((ax_u8)('/'))) || ((*((ax_u8*)(ch))) == ((ax_u8)('\\'))))) {
            last_sep = i;
        } else if (((*((ax_u8*)(ch))) == ((ax_u8)('.')))) {
            last_dot = i;
        }
        i = (i + 1);
    }
    if ((last_dot > last_sep)) {
        return ax_pathbuf_new(ax_str_concat(ax_str_concat(ax_str_slice(s, 0, last_dot), (ax_string){.ptr=(const ax_u8*)".", .len=1}), ext));
    }
    return ax_pathbuf_new(ax_str_concat(ax_str_concat(s, (ax_string){.ptr=(const ax_u8*)".", .len=1}), ext));
}

ax_bool ax_PathBuf_pathbuf_is_absolute(struct ax_PathBuf self) {
    ax_string s = self.inner;
    if ((s.len == 0)) {
        return AX_FALSE;
    }
    if (ax_IS_WINDOWS) {
        if ((s.len >= 2)) {
            ax_u8* c1 = ((ax_u8*)((((ax_i64)(s.ptr)) + 0)));
            ax_u8* c2 = ((ax_u8*)((((ax_i64)(s.ptr)) + 1)));
            ax_bool is_letter = ((((*((ax_u8*)(c1))) >= ((ax_u8)('a'))) && ((*((ax_u8*)(c1))) <= ((ax_u8)('z')))) || (((*((ax_u8*)(c1))) >= ((ax_u8)('A'))) && ((*((ax_u8*)(c1))) <= ((ax_u8)('Z')))));
            if ((is_letter && ((*((ax_u8*)(c2))) == ((ax_u8)(':'))))) {
                return AX_TRUE;
            }
        }
        return AX_FALSE;
    } else {
        {
            ax_u8* c1 = ((ax_u8*)(s.ptr));
            return ((*((ax_u8*)(c1))) == ((ax_u8)('/')));
        }
    }
}

ax_string ax_PathBuf_pathbuf_to_str(struct ax_PathBuf self) {
    return self.inner;
}

static ax_u8* ax_os_str_to_c_str(ax_string s) {
    ax_i64 len = s.len;
    ax_u8* buf = ((ax_u8*)(ax_alloc((len + 1))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < len)) {
        (*((ax_u8*)(((ax_u8*)((((ax_i64)(buf)) + i)))))) = (*((ax_u8*)(((ax_u8*)((((ax_i64)(s.ptr)) + i))))));
        i = (i + 1);
    }
    (*((ax_u8*)(((ax_u8*)((((ax_i64)(buf)) + len)))))) = ((ax_u8)(0));
    return buf;
    ax_free(buf);
}

struct ax__AX_std_Result__FileMetadata__string ax_metadata(ax_string path) {
    ax_u8* cpath = ax_os_str_to_c_str(path);
    if (ax_IS_WINDOWS) {
        ax_u32 attrs = GetFileAttributesA(cpath);
        ax_free(cpath);
        if ((attrs == ((ax_u32)(0xFFFFFFFF)))) {
            return ax__AX_std_Result__FileMetadata__string_err((ax_string){.ptr=(const ax_u8*)"file not found", .len=14});
        }
        ax_bool is_dir = ((attrs & ((ax_u32)(0x10))) != ((ax_u32)(0)));
        ax_bool is_symlink = ((attrs & ((ax_u32)(0x400))) != ((ax_u32)(0)));
        return ax__AX_std_Result__FileMetadata__string_ok(((struct ax_FileMetadata){.size=((ax_u64)(0)), .modified=((ax_u64)(0)), .created=((ax_u64)(0)), .is_file=(!is_dir), .is_dir=is_dir, .is_symlink=is_symlink, .mode=((ax_u32)(0))}));
    } else {
        {
            ax_u8* stat_buf = ((ax_u8*)(ax_alloc(144)));
            ax_i64 res = syscall(((ax_u64)(4)), ((ax_u64)(cpath)), ((ax_u64)(stat_buf)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            ax_free(cpath);
            if ((res < 0)) {
                ax_free(stat_buf);
                return ax__AX_std_Result__FileMetadata__string_err((ax_string){.ptr=(const ax_u8*)"file not found", .len=14});
            }
            ax_u32 st_mode = (*((ax_u32*)(((ax_u32*)((((ax_i64)(stat_buf)) + 24))))));
            ax_u64 st_size = (*((ax_u64*)(((ax_u64*)((((ax_i64)(stat_buf)) + 48))))));
            ax_u64 st_mtime = (*((ax_u64*)(((ax_u64*)((((ax_i64)(stat_buf)) + 88))))));
            ax_bool is_dir = ((st_mode & ((ax_u32)(0xF000))) == ((ax_u32)(0x4000)));
            ax_bool is_lnk = ((st_mode & ((ax_u32)(0xF000))) == ((ax_u32)(0xA000)));
            ax_free(stat_buf);
            return ax__AX_std_Result__FileMetadata__string_ok(((struct ax_FileMetadata){.size=st_size, .modified=(st_mtime * ((ax_u64)(1000000000))), .created=((ax_u64)(0)), .is_file=((!is_dir) && (!is_lnk)), .is_dir=is_dir, .is_symlink=is_lnk, .mode=st_mode}));
        }
    }
}

ax_bool ax_exists(ax_string path) {
    return ax_AX_std_is_ok__FileMetadata__string(ax_metadata(path));
}

struct ax__AX_std_Result__void__string ax_create_dir(ax_string path) {
    ax_u8* cpath = ax_os_str_to_c_str(path);
    if (ax_IS_WINDOWS) {
        ax_i32 res = CreateDirectoryA(cpath, ((void*)(NULL)));
        ax_free(cpath);
        if ((res == 0)) {
            return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not create directory", .len=26});
        }
        return ax__AX_std_Result__void__string_ok();
    } else {
        {
            ax_i64 res = syscall(((ax_u64)(83)), ((ax_u64)(cpath)), ((ax_u64)(0o777)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            ax_free(cpath);
            if ((res < 0)) {
                return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not create directory", .len=26});
            }
            return ax__AX_std_Result__void__string_ok();
        }
    }
}

struct ax__AX_std_Result__void__string ax_create_dir_all(ax_string path) {
    if (ax_exists(path)) {
        return ax__AX_std_Result__void__string_ok();
    }
    struct ax_PathBuf p = ax_pathbuf_new(path);
    struct ax__AX_std_Option__PathBuf parent_opt = ax_PathBuf_pathbuf_parent(p);
    if (ax_AX_std_is_some__PathBuf(parent_opt)) {
        struct ax_PathBuf parent = ax_AX_std_unwrap__PathBuf(parent_opt);
        ax_string parent_str = ax_PathBuf_pathbuf_to_str(parent);
        if ((parent_str.len > 0)) {
            struct ax__AX_std_Result__void__string res = ax_create_dir_all(parent_str);
            if (ax_AX_std_is_err__void__string(res)) {
                return res;
            }
            /* skip destroy for value type res */
        }
    }
    return ax_create_dir(path);
}

struct ax__AX_std_Result__void__string ax_remove_file(ax_string path) {
    ax_u8* cpath = ax_os_str_to_c_str(path);
    if (ax_IS_WINDOWS) {
        ax_i32 res = DeleteFileA(cpath);
        ax_free(cpath);
        if ((res == 0)) {
            return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not remove file", .len=21});
        }
        return ax__AX_std_Result__void__string_ok();
    } else {
        {
            ax_i64 res = syscall(((ax_u64)(87)), ((ax_u64)(cpath)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            ax_free(cpath);
            if ((res < 0)) {
                return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not remove file", .len=21});
            }
            return ax__AX_std_Result__void__string_ok();
        }
    }
}

struct ax__AX_std_Result__void__string ax_remove_dir(ax_string path) {
    ax_u8* cpath = ax_os_str_to_c_str(path);
    if (ax_IS_WINDOWS) {
        ax_i32 res = RemoveDirectoryA(cpath);
        ax_free(cpath);
        if ((res == 0)) {
            return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not remove directory", .len=26});
        }
        return ax__AX_std_Result__void__string_ok();
    } else {
        {
            ax_i64 res = syscall(((ax_u64)(84)), ((ax_u64)(cpath)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            ax_free(cpath);
            if ((res < 0)) {
                return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not remove directory", .len=26});
            }
            return ax__AX_std_Result__void__string_ok();
        }
    }
}

struct ax__AX_std_Result__void__string ax_rename(ax_string from, ax_string to) {
    ax_u8* cfrom = ax_os_str_to_c_str(from);
    ax_u8* cto = ax_os_str_to_c_str(to);
    if (ax_IS_WINDOWS) {
        ax_i32 res = MoveFileA(cfrom, cto);
        ax_free(cfrom);
        ax_free(cto);
        if ((res == 0)) {
            return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not rename", .len=16});
        }
        return ax__AX_std_Result__void__string_ok();
    } else {
        {
            ax_i64 res = syscall(((ax_u64)(82)), ((ax_u64)(cfrom)), ((ax_u64)(cto)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            ax_free(cfrom);
            ax_free(cto);
            if ((res < 0)) {
                return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"could not rename", .len=16});
            }
            return ax__AX_std_Result__void__string_ok();
        }
    }
}

struct ax__AX_std_Result__u64__string ax_copy(ax_string from, ax_string to) {
    ax_u8* cfrom = ax_os_str_to_c_str(from);
    ax_u8* cto = ax_os_str_to_c_str(to);
    if (ax_IS_WINDOWS) {
        ax_i32 res = CopyFileA(cfrom, cto, ((ax_i32)(0)));
        ax_free(cfrom);
        ax_free(cto);
        if ((res == 0)) {
            return ax__AX_std_Result__u64__string_err((ax_string){.ptr=(const ax_u8*)"could not copy file", .len=19});
        }
        return ax__AX_std_Result__u64__string_ok(((ax_u64)(0)));
    } else {
        {
            ax_free(cfrom);
            ax_free(cto);
            return ax__AX_std_Result__u64__string_err((ax_string){.ptr=(const ax_u8*)"copy not implemented on Linux in freestanding mode", .len=50});
        }
    }
}

struct ax_PathBuf ax_temp_dir(void) {
    struct ax__AX_std_Option__string t = ax_env((ax_string){.ptr=(const ax_u8*)"TMPDIR", .len=6});
    if (ax_AX_std_is_some__string(t)) {
        return ax_pathbuf_new(ax_AX_std_unwrap__string(t));
    }
    struct ax__AX_std_Option__string t2 = ax_env((ax_string){.ptr=(const ax_u8*)"TEMP", .len=4});
    if (ax_AX_std_is_some__string(t2)) {
        return ax_pathbuf_new(ax_AX_std_unwrap__string(t2));
    }
    if (ax_IS_WINDOWS) {
        return ax_pathbuf_new((ax_string){.ptr=(const ax_u8*)"C:\\Windows\\Temp", .len=15});
    }
    return ax_pathbuf_new((ax_string){.ptr=(const ax_u8*)"/tmp", .len=4});
}

void ax_trap_signal(ax_i32 sig, void (*handler)(void)) {
    return;
}

ax_vec ax_parse_cmdline(ax_u8* cmdline) {
    ax_vec res = ax_vec_new(sizeof(ax_string));
    if ((cmdline == ((ax_u8*)(NULL)))) {
        return res;
    }
    ax_i64 i = ((ax_i64)(0));
    while (((*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i)))))) != ((ax_u8)(0)))) {
        while ((((((*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i)))))) == ((ax_u8)(' '))) || ((*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i)))))) == ((ax_u8)('\t')))) || ((*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i)))))) == ((ax_u8)('\r')))) || ((*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i)))))) == ((ax_u8)('\n'))))) {
            i = (i + 1);
        }
        if (((*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i)))))) == ((ax_u8)(0)))) {
            break;
        }
        ax_vec arg_buf = ax_vec_new(sizeof(ax_u8));
        ax_bool in_quotes = AX_FALSE;
        while (((*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i)))))) != ((ax_u8)(0)))) {
            ax_u8 ch = (*((ax_u8*)(((ax_u8*)((((ax_i64)(cmdline)) + i))))));
            if ((ch == ((ax_u8)('"')))) {
                in_quotes = (!in_quotes);
            } else if ((((ch == ((ax_u8)(' '))) || (ch == ((ax_u8)('\t')))) && (!in_quotes))) {
                break;
            } else {
                {
                    ax__AX_std_Vec__u8_push(&(arg_buf), ch);
                }
            }
            i = (i + 1);
        }
        ax_i64 arg_len = arg_buf.len;
        ax_u8* s_ptr = ((ax_u8*)(ax_alloc((arg_len + 1))));
        ax_i64 j = ((ax_i64)(0));
        while ((j < arg_len)) {
            (((ax_u8*)(s_ptr))[j]) = (((ax_u8*)(arg_buf.data))[j]);
            j = (j + 1);
        }
        (((ax_u8*)(s_ptr))[arg_len]) = ((ax_u8)(0));
        ax__AX_std_Vec__string_push(&(res), ((ax_string){.ptr = (const ax_u8*)(s_ptr), .len = strlen((const char*)(s_ptr))}));
        ax__AX_std_Vec__u8_destroy(&(arg_buf));
    }
    return res;
    /* skip destroy for value type res */
}

ax_vec ax_parse_cmdline_w(ax_u16* cmdline) {
    ax_vec res = ax_vec_new(sizeof(ax_string));
    if ((cmdline == ((ax_u16*)(NULL)))) {
        return res;
    }
    ax_i64 i = ((ax_i64)(0));
    while (((((ax_u16*)(cmdline))[i]) != ((ax_u16)(0)))) {
        while ((((((((ax_u16*)(cmdline))[i]) == ((ax_u16)(' '))) || ((((ax_u16*)(cmdline))[i]) == ((ax_u16)('\t')))) || ((((ax_u16*)(cmdline))[i]) == ((ax_u16)('\r')))) || ((((ax_u16*)(cmdline))[i]) == ((ax_u16)('\n'))))) {
            i = (i + 1);
        }
        if (((((ax_u16*)(cmdline))[i]) == ((ax_u16)(0)))) {
            break;
        }
        ax_vec arg_buf = ax_vec_new(sizeof(ax_u16));
        ax_bool in_quotes = AX_FALSE;
        while (((((ax_u16*)(cmdline))[i]) != ((ax_u16)(0)))) {
            ax_u16 ch = (((ax_u16*)(cmdline))[i]);
            if ((ch == ((ax_u16)('"')))) {
                in_quotes = (!in_quotes);
            } else if ((((ch == ((ax_u16)(' '))) || (ch == ((ax_u16)('\t')))) && (!in_quotes))) {
                break;
            } else {
                {
                    ax__AX_std_Vec__u16_push(&(arg_buf), ch);
                }
            }
            i = (i + 1);
        }
        ax_i64 arg_len = arg_buf.len;
        ax_u8* s_ptr = ((ax_u8*)(ax_alloc(((arg_len * 3) + 4))));
        ax_i64 j = ((ax_i64)(0));
        ax_i64 out_idx = ((ax_i64)(0));
        while ((j < arg_len)) {
            ax_u16 w1 = (((ax_u16*)(((ax_u16*)(arg_buf.data))))[j]);
            ax_i64 cp = ((ax_i64)(w1));
            if ((((w1 >= ((ax_u16)(0xD800))) && (w1 <= ((ax_u16)(0xDBFF)))) && ((j + 1) < arg_len))) {
                ax_u16 w2 = (((ax_u16*)(((ax_u16*)(arg_buf.data))))[(j + 1)]);
                if (((w2 >= ((ax_u16)(0xDC00))) && (w2 <= ((ax_u16)(0xDFFF))))) {
                    cp = (((((ax_i64)((w1 - ((ax_u16)(0xD800))))) << 10) + ((ax_i64)((w2 - ((ax_u16)(0xDC00)))))) + ((ax_i64)(65536)));
                    j = (j + 1);
                }
            }
            if ((cp < ((ax_i64)(128)))) {
                (((ax_u8*)(s_ptr))[out_idx]) = ((ax_u8)(cp));
                out_idx = (out_idx + 1);
            } else if ((cp < ((ax_i64)(2048)))) {
                (((ax_u8*)(s_ptr))[out_idx]) = ((ax_u8)((((ax_i64)(192)) + (cp >> 6))));
                (((ax_u8*)(s_ptr))[(out_idx + 1)]) = ((ax_u8)((((ax_i64)(128)) + (cp & ((ax_i64)(63))))));
                out_idx = (out_idx + 2);
            } else if ((cp < ((ax_i64)(65536)))) {
                (((ax_u8*)(s_ptr))[out_idx]) = ((ax_u8)((((ax_i64)(224)) + (cp >> 12))));
                (((ax_u8*)(s_ptr))[(out_idx + 1)]) = ((ax_u8)((((ax_i64)(128)) + ((cp >> 6) & ((ax_i64)(63))))));
                (((ax_u8*)(s_ptr))[(out_idx + 2)]) = ((ax_u8)((((ax_i64)(128)) + (cp & ((ax_i64)(63))))));
                out_idx = (out_idx + 3);
            } else {
                {
                    (((ax_u8*)(s_ptr))[out_idx]) = ((ax_u8)((((ax_i64)(240)) + (cp >> 18))));
                    (((ax_u8*)(s_ptr))[(out_idx + 1)]) = ((ax_u8)((((ax_i64)(128)) + ((cp >> 12) & ((ax_i64)(63))))));
                    (((ax_u8*)(s_ptr))[(out_idx + 2)]) = ((ax_u8)((((ax_i64)(128)) + ((cp >> 6) & ((ax_i64)(63))))));
                    (((ax_u8*)(s_ptr))[(out_idx + 3)]) = ((ax_u8)((((ax_i64)(128)) + (cp & ((ax_i64)(63))))));
                    out_idx = (out_idx + 4);
                }
            }
            j = (j + 1);
        }
        (((ax_u8*)(s_ptr))[out_idx]) = ((ax_u8)(0));
        ax__AX_std_Vec__string_push(&(res), ((ax_string){.ptr = (const ax_u8*)(s_ptr), .len = strlen((const char*)(s_ptr))}));
        ax__AX_std_Vec__u16_destroy(&(arg_buf));
    }
    return res;
    /* skip destroy for value type res */
}

ax_vec ax_parse_proc_cmdline(ax_u8* buf, ax_i64 len) {
    ax_vec res = ax_vec_new(sizeof(ax_string));
    ax_i64 start = ((ax_i64)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < len)) {
        if (((((ax_u8*)(buf))[i]) == ((ax_u8)(0)))) {
            if ((i > start)) {
                ax_i64 arg_len = (i - start);
                ax_u8* s_ptr = ((ax_u8*)(ax_alloc((arg_len + 1))));
                memcpy(s_ptr, ((ax_u8*)((((ax_i64)(buf)) + start))), arg_len);
                (((ax_u8*)(s_ptr))[arg_len]) = ((ax_u8)(0));
                ax__AX_std_Vec__string_push(&(res), ((ax_string){.ptr = (const ax_u8*)(s_ptr), .len = strlen((const char*)(s_ptr))}));
            }
            start = (i + 1);
        }
        i = (i + 1);
    }
    return res;
    /* skip destroy for value type res */
}

static ax_i64 ax_linux_read_file(ax_string path, ax_u8* buf, ax_i64 max_len) {
    ax_u8* cpath = ax_os_str_to_c_str(path);
    ax_i64 fd = syscall(((ax_u64)(2)), ((ax_u64)(cpath)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
    ax_free(cpath);
    if ((fd < 0)) {
        return (-1);
    }
    ax_i64 n = syscall(((ax_u64)(0)), ((ax_u64)(fd)), ((ax_u64)(buf)), ((ax_u64)(max_len)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
    syscall(((ax_u64)(3)), ((ax_u64)(fd)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
    return n;
}

ax_vec ax_args(void) {
    if (ax_IS_WINDOWS) {
        ax_u16* cmdline = GetCommandLineW();
        return ax_parse_cmdline_w(cmdline);
    } else {
        {
            ax_u8* buf = ((ax_u8*)(ax_alloc(8192)));
            ax_i64 n = ax_linux_read_file((ax_string){.ptr=(const ax_u8*)"/proc/self/cmdline", .len=18}, buf, 8192);
            if ((n < 0)) {
                ax_free(buf);
                return ax_vec_new(sizeof(ax_string));
            }
            ax_vec res = ax_parse_proc_cmdline(buf, n);
            ax_free(buf);
            return res;
            /* skip destroy for value type res */
        }
    }
}

struct ax__AX_std_Option__string ax_env(ax_string key) {
    if (ax_IS_WINDOWS) {
        ax_u8* ckey = ax_os_str_to_c_str(key);
        ax_u8* buf = ((ax_u8*)(ax_alloc(4096)));
        ax_u32 n = GetEnvironmentVariableA(ckey, buf, ((ax_u32)(4096)));
        ax_free(ckey);
        if ((n == ((ax_u32)(0)))) {
            ax_free(buf);
            return ax__AX_std_Option__string_none();
        }
        ax_string res_str = ax_str_slice(((ax_string){.ptr = (const ax_u8*)(buf), .len = strlen((const char*)(buf))}), 0, ((ax_i64)(n)));
        ax_free(buf);
        return ax__AX_std_Option__string_some(res_str);
    } else {
        {
            ax_u8* buf = ((ax_u8*)(ax_alloc(16384)));
            ax_i64 n = ax_linux_read_file((ax_string){.ptr=(const ax_u8*)"/proc/self/environ", .len=18}, buf, 16384);
            if ((n < 0)) {
                ax_free(buf);
                return ax__AX_std_Option__string_none();
            }
            ax_string prefix = ax_str_concat(key, (ax_string){.ptr=(const ax_u8*)"=", .len=1});
            ax_i64 start = ((ax_i64)(0));
            ax_i64 i = ((ax_i64)(0));
            while ((i < n)) {
                if (((((ax_u8*)(buf))[i]) == ((ax_u8)(0)))) {
                    if ((i > start)) {
                        ax_string entry = ax_str_slice(((ax_string){.ptr = (const ax_u8*)(buf), .len = strlen((const char*)(buf))}), start, i);
                        if (ax_std_string_starts_with(entry, prefix)) {
                            ax_string val = ax_str_slice(entry, prefix.len, entry.len);
                            ax_free(buf);
                            return ax__AX_std_Option__string_some(val);
                        }
                    }
                    start = (i + 1);
                }
                i = (i + 1);
            }
            ax_free(buf);
            return ax__AX_std_Option__string_none();
        }
    }
}

struct ax_Command ax_new_command(ax_string program) {
    return ((struct ax_Command){.program=program, .args=ax_AX_std_new_vec__string(), .env=ax_AX_std_new_hashmap__string__string(), .cwd=ax__AX_std_Option__string_none()});
}

void ax_Command_arg(struct ax_Command* self, ax_string arg) {
    ax__AX_std_Vec__string_push(&(self->args), arg);
}

void ax_Command_env(struct ax_Command* self, ax_string key, ax_string value) {
    ax__AX_std_HashMap__string__string_insert(&(self->env), key, value);
}

void ax_Command_cwd(struct ax_Command* self, ax_string dir) {
    self->cwd = ax__AX_std_Option__string_some(dir);
}

struct ax__AX_std_Result__Child__string ax_Command_spawn_process(struct ax_Command self) {
    if (1) {
        void* si = ((void*)(ax_alloc(104)));
        memset(((ax_u8*)(si)), ((ax_u8)(0)), ((ax_i64)(104)));
        (((ax_u32*)(((ax_u32*)(si))))[0]) = ((ax_u32)(104));
        void* pi = ((void*)(ax_alloc(24)));
        memset(((ax_u8*)(pi)), ((ax_u8)(0)), ((ax_i64)(24)));
        ax_string cmdline = self.program;
        ax_i64 i = ((ax_i64)(0));
        while ((i < self.args.len)) {
            ax_string arg = ax_AX_std_unwrap__string(ax__AX_std_Vec__string_get(self.args, i));
            cmdline = ax_str_concat(cmdline, ax_str_concat((ax_string){.ptr=(const ax_u8*)" ", .len=1}, arg));
            i = (i + 1);
        }
        ax_u8* cmdline_nt = ((ax_u8*)(ax_alloc((ax_str_len(cmdline) + 1))));
        memcpy(cmdline_nt, ((ax_u8*)(cmdline.ptr)), ax_str_len(cmdline));
        (((ax_u8*)(cmdline_nt))[ax_str_len(cmdline)]) = ((ax_u8)(0));
        ax_i32 res = CreateProcessA(((void*)(NULL)), ((void*)(cmdline_nt)), ((void*)(NULL)), ((void*)(NULL)), ((ax_i32)(1)), ((ax_u32)(0)), ((void*)(NULL)), ((void*)(NULL)), si, pi);
        ax_free(cmdline_nt);
        ax_free(((ax_u8*)(si)));
        if ((res == 0)) {
            ax_free(((ax_u8*)(pi)));
            return ax__AX_std_Result__Child__string_err((ax_string){.ptr=(const ax_u8*)"failed to spawn child process via CreateProcessA", .len=48});
        }
        void* hProcess = (((void**)(((void**)(pi))))[0]);
        void* hThread = (((void**)(((void**)(pi))))[1]);
        CloseHandle(hThread);
        ax_free(((ax_u8*)(pi)));
        return ax__AX_std_Result__Child__string_ok(((struct ax_Child){.pid=((ax_i64)(hProcess))}));
    } else {
        {
            ax_i64 pid = syscall(((ax_u64)(57)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            if ((pid < 0)) {
                return ax__AX_std_Result__Child__string_err((ax_string){.ptr=(const ax_u8*)"fork failed", .len=11});
            }
            if ((pid == 0)) {
                ax_u8* prog_nt = ((ax_u8*)(ax_alloc((ax_str_len(self.program) + 1))));
                memcpy(prog_nt, ((ax_u8*)(self.program.ptr)), ax_str_len(self.program));
                (((ax_u8*)(prog_nt))[ax_str_len(self.program)]) = ((ax_u8)(0));
                ax_u8** argv = ((ax_u8**)(ax_alloc(((self.args.len + 2) * 8))));
                (((ax_u8**)(argv))[0]) = prog_nt;
                ax_i64 idx = ((ax_i64)(0));
                while ((idx < self.args.len)) {
                    ax_string arg = ax_AX_std_unwrap__string(ax__AX_std_Vec__string_get(self.args, idx));
                    ax_u8* arg_nt = ((ax_u8*)(ax_alloc((ax_str_len(arg) + 1))));
                    memcpy(arg_nt, ((ax_u8*)(arg.ptr)), ax_str_len(arg));
                    (((ax_u8*)(arg_nt))[ax_str_len(arg)]) = ((ax_u8)(0));
                    (((ax_u8**)(argv))[(idx + 1)]) = arg_nt;
                    idx = (idx + 1);
                }
                (((ax_u8**)(argv))[(self.args.len + 1)]) = ((ax_u8*)(NULL));
                syscall(((ax_u64)(59)), ((ax_u64)(prog_nt)), ((ax_u64)(argv)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
                syscall(((ax_u64)(60)), ((ax_u64)(127)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            }
            return ax__AX_std_Result__Child__string_ok(((struct ax_Child){.pid=pid}));
        }
    }
}

struct ax__AX_std_Result__Output__string ax_Command_output(struct ax_Command self) {
    {
        struct ax__AX_std_Result__Child__string _discrim = ax_Command_spawn_process(self);
        switch (_discrim.tag) {
        case ax__AX_std_Result__Child__string_Ok: {
            struct ax_Child child = (_discrim).data.Ok;
            struct ax_Child c = child;
            {
                struct ax__AX_std_Result__ExitStatus__string _discrim = ax_Child_wait(&(c));
                switch (_discrim.tag) {
                case ax__AX_std_Result__ExitStatus__string_Ok: {
                    struct ax_ExitStatus status = (_discrim).data.Ok;
                    return ax__AX_std_Result__Output__string_ok(((struct ax_Output){.status=status, .stdout_str=(ax_string){.ptr=(const ax_u8*)"", .len=0}, .stderr_str=(ax_string){.ptr=(const ax_u8*)"", .len=0}}));
                    break;
                }
                case ax__AX_std_Result__ExitStatus__string_Err: {
                    ax_string e = (_discrim).data.Err;
                    return ax__AX_std_Result__Output__string_err(e);
                    break;
                }
                    default: {
                        /* unreachable: exhaustiveness checked by type checker */
                        __builtin_unreachable();
                    }
                }
            }
            break;
        }
        case ax__AX_std_Result__Child__string_Err: {
            ax_string e = (_discrim).data.Err;
            return ax__AX_std_Result__Output__string_err(e);
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
}

struct ax__AX_std_Result__ExitStatus__string ax_Child_wait(struct ax_Child* self) {
    if (1) {
        void* hProcess = ((void*)(self->pid));
        ax_u32 wait_res = WaitForSingleObject(hProcess, ((ax_u32)(0xFFFFFFFF)));
        ax_u32 exit_code = ((ax_u32)(0));
        ax_i32 ok = GetExitCodeProcess(hProcess, ((void*)(&(exit_code))));
        CloseHandle(hProcess);
        if ((ok == 0)) {
            return ax__AX_std_Result__ExitStatus__string_err((ax_string){.ptr=(const ax_u8*)"failed to get child exit code", .len=29});
        }
        return ax__AX_std_Result__ExitStatus__string_ok(((struct ax_ExitStatus){.code=((ax_i32)(exit_code))}));
    } else {
        {
            ax_i32 status = ((ax_i32)(0));
            ax_i64 res = syscall(((ax_u64)(61)), ((ax_u64)(self->pid)), ((ax_u64)(((ax_i64)(&(status))))), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            if ((res < 0)) {
                return ax__AX_std_Result__ExitStatus__string_err((ax_string){.ptr=(const ax_u8*)"waitpid failed", .len=14});
            }
            ax_i32 exit_code = ((status >> 8) & 0xFF);
            return ax__AX_std_Result__ExitStatus__string_ok(((struct ax_ExitStatus){.code=exit_code}));
        }
    }
}

struct ax__AX_std_Result__void__string ax_Child_kill(struct ax_Child self) {
    if (1) {
        void* hProcess = ((void*)(self.pid));
        ax_i32 ok = TerminateProcess(hProcess, ((ax_u32)(1)));
        CloseHandle(hProcess);
        if ((ok == 0)) {
            return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"failed to terminate process", .len=27});
        }
        return ax__AX_std_Result__void__string_ok();
    } else {
        {
            ax_i64 res = syscall(((ax_u64)(62)), ((ax_u64)(self.pid)), ((ax_u64)(9)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            if ((res < 0)) {
                return ax__AX_std_Result__void__string_err((ax_string){.ptr=(const ax_u8*)"failed to kill process", .len=22});
            }
            return ax__AX_std_Result__void__string_ok();
        }
    }
}

ax_bool ax_ExitStatus_success(struct ax_ExitStatus self) {
    return (self.code == 0);
}

ax_i32 ax_ExitStatus_code(struct ax_ExitStatus self) {
    return self.code;
}

void exit(ax_i32 code) {
    if (1) {
        ExitProcess(((ax_u32)(code)));
    } else {
        {
            syscall(((ax_u64)(60)), ((ax_u64)(code)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

void abort(void) {
    if (1) {
        ExitProcess(((ax_u32)(101)));
    } else {
        {
            ax_i64 pid = syscall(((ax_u64)(39)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            syscall(((ax_u64)(62)), ((ax_u64)(pid)), ((ax_u64)(6)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

static void ax_test_print_str(ax_string s) {
    if (1) {
        void* h = GetStdHandle(((ax_u32)(0xFFFFFFF5)));
        ax_u32 written = ((ax_u32)(0));
        WriteFile(h, ((void*)(s.ptr)), ((ax_u32)(ax_str_len(s))), ((void*)(&(written))), ((void*)(NULL)));
    } else {
        {
            syscall(((ax_u64)(1)), ((ax_u64)(1)), ((ax_u64)(((ax_u8*)(s.ptr)))), ((ax_u64)(ax_str_len(s))), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

static void ax_test_process_spawn(void) {
    struct ax_Command cmd = ax_new_command((ax_string){.ptr=(const ax_u8*)"cmd.exe", .len=7});
    if ((!1)) {
        cmd = ax_new_command((ax_string){.ptr=(const ax_u8*)"/bin/echo", .len=9});
    }
    if (1) {
        ax_Command_arg(&(cmd), (ax_string){.ptr=(const ax_u8*)"/c", .len=2});
        ax_Command_arg(&(cmd), (ax_string){.ptr=(const ax_u8*)"echo", .len=4});
        ax_Command_arg(&(cmd), (ax_string){.ptr=(const ax_u8*)"Process_Test_Successful", .len=23});
    } else {
        {
            ax_Command_arg(&(cmd), (ax_string){.ptr=(const ax_u8*)"Process_Test_Successful", .len=23});
        }
    }
    {
        struct ax__AX_std_Result__Output__string _discrim = ax_Command_output(cmd);
        switch (_discrim.tag) {
        case ax__AX_std_Result__Output__string_Ok: {
            struct ax_Output output = (_discrim).data.Ok;
            ax_assert_axiom((ax_ExitStatus_success(output.status) == AX_TRUE), AX_STR("(ax_ExitStatus_success(output.status) == AX_TRUE)"));
            ax_assert_axiom((output.status.code == 0), AX_STR("(output.status.code == 0)"));
            break;
        }
        case ax__AX_std_Result__Output__string_Err: {
            ax_string e = (_discrim).data.Err;
            ax_test_print_str((ax_string){.ptr=(const ax_u8*)"Error: ", .len=7});
            ax_test_print_str(e);
            ax_test_print_str((ax_string){.ptr=(const ax_u8*)"\n", .len=1});
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_process_spawn\n", .len=27});
}

ax_i32 ax_main_usr(void) {
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"Running AXIOM-native process unit tests...\n", .len=43});
    ax_test_process_spawn();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"All AXIOM-native process tests passed!\n", .len=39});
    return 0;
}

ax_bool ax_AX_std_is_ok__FileMetadata__string(struct ax__AX_std_Result__FileMetadata__string self) {
    if (ax_sum_layout_is_pointer()) {
        ax_u64* raw = ((ax_u64*)(&(self)));
        return (((((ax_u64*)(raw))[0]) & ((ax_u64)(1))) == ((ax_u64)(0)));
    }
    ax_u64 size = (sizeof(struct ax_FileMetadata) * ((ax_u64)(2)));
    if ((size <= ((ax_u64)(8)))) {
        return AX_TRUE;
    }
    ax_u32* raw = ((ax_u32*)(&(self)));
    return ((((ax_u32*)(raw))[0]) == ((ax_u32)(0)));
}

ax_bool ax_AX_std_is_some__PathBuf(struct ax__AX_std_Option__PathBuf self) {
    if (ax_sum_layout_is_pointer()) {
        ax_u64* raw = ((ax_u64*)(&(self)));
        return ((((ax_u64*)(raw))[0]) != ((ax_u64)(0)));
    }
    ax_u64 size = (sizeof(struct ax_PathBuf) * ((ax_u64)(2)));
    if ((size <= ((ax_u64)(8)))) {
        ax_u64* raw = ((ax_u64*)(&(self)));
        return ((((ax_u64*)(raw))[0]) != ((ax_u64)(0)));
    }
    ax_u32* raw = ((ax_u32*)(&(self)));
    return ((((ax_u32*)(raw))[0]) == ((ax_u32)(0)));
}

struct ax_PathBuf ax_AX_std_unwrap__PathBuf(struct ax__AX_std_Option__PathBuf self) {
    if ((!ax_AX_std_is_some__PathBuf(self))) {
        ax_panic((const char*)((ax_string){.ptr=(const ax_u8*)"called Option.unwrap() on a None value", .len=38}).ptr);
    }
    if (ax_sum_layout_is_pointer()) {
        struct ax_PathBuf* raw = ((struct ax_PathBuf*)(&(self)));
        return (*((struct ax_PathBuf*)(raw)));
    }
    ax_u64 size = (sizeof(struct ax_PathBuf) * ((ax_u64)(2)));
    if ((size <= ((ax_u64)(8)))) {
        ax_u8* raw = ((ax_u8*)(&(self)));
        struct ax_PathBuf* typed = ((struct ax_PathBuf*)(raw));
        return (*((struct ax_PathBuf*)(typed)));
    }
    ax_u8* raw = ((ax_u8*)(&(self)));
    struct ax_PathBuf* payload_ptr = ((struct ax_PathBuf*)((((ax_i64)(raw)) + ((ax_i64)(8)))));
    return (*((struct ax_PathBuf*)(payload_ptr)));
}

ax_bool ax_AX_std_is_err__void__string(struct ax__AX_std_Result__void__string self) {
    if (ax_sum_layout_is_pointer()) {
        ax_u64* raw = ((ax_u64*)(&(self)));
        return (((((ax_u64*)(raw))[0]) & ((ax_u64)(1))) == ((ax_u64)(1)));
    }
    ax_u64 size = (sizeof(void) * ((ax_u64)(2)));
    if ((size <= ((ax_u64)(8)))) {
        return AX_FALSE;
    }
    ax_u32* raw = ((ax_u32*)(&(self)));
    return ((((ax_u32*)(raw))[0]) == ((ax_u32)(1)));
}

ax_bool ax_AX_std_is_some__string(struct ax__AX_std_Option__string self) {
    if (ax_sum_layout_is_pointer()) {
        ax_u64* raw = ((ax_u64*)(&(self)));
        return ((((ax_u64*)(raw))[0]) != ((ax_u64)(0)));
    }
    ax_u64 size = (sizeof(ax_string) * ((ax_u64)(2)));
    if ((size <= ((ax_u64)(8)))) {
        ax_u64* raw = ((ax_u64*)(&(self)));
        return ((((ax_u64*)(raw))[0]) != ((ax_u64)(0)));
    }
    ax_u32* raw = ((ax_u32*)(&(self)));
    return ((((ax_u32*)(raw))[0]) == ((ax_u32)(0)));
}

ax_string ax_AX_std_unwrap__string(struct ax__AX_std_Option__string self) {
    if ((!ax_AX_std_is_some__string(self))) {
        ax_panic((const char*)((ax_string){.ptr=(const ax_u8*)"called Option.unwrap() on a None value", .len=38}).ptr);
    }
    if (ax_sum_layout_is_pointer()) {
        ax_string* raw = ((ax_string*)(&(self)));
        return (*((ax_string*)(raw)));
    }
    ax_u64 size = (sizeof(ax_string) * ((ax_u64)(2)));
    if ((size <= ((ax_u64)(8)))) {
        ax_u8* raw = ((ax_u8*)(&(self)));
        ax_string* typed = ((ax_string*)(raw));
        return (*((ax_string*)(typed)));
    }
    ax_u8* raw = ((ax_u8*)(&(self)));
    ax_string* payload_ptr = ((ax_string*)((((ax_i64)(raw)) + ((ax_i64)(8)))));
    return (*((ax_string*)(payload_ptr)));
}

void ax__AX_std_Vec__u8_push(ax_vec* self, ax_u8 item) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_i64 item_size = ((ax_i64)(sizeof(ax_u8)));
        ax_u8* new_data = ((ax_u8*)(ax_alloc((new_cap * item_size))));
        if ((self->data != ((ax_u8*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * item_size));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    (((ax_u8*)(self->data))[self->len]) = item;
    self->len = (self->len + 1);
}

void ax__AX_std_Vec__string_push(ax_vec* self, ax_string item) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_i64 item_size = ((ax_i64)(sizeof(ax_string)));
        ax_string* new_data = ((ax_string*)(ax_alloc((new_cap * item_size))));
        if ((self->data != ((ax_string*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * item_size));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    (((ax_string*)(self->data))[self->len]) = item;
    self->len = (self->len + 1);
}

void ax__AX_std_Vec__u8_destroy(ax_vec* self) {
    if ((self->data != ((ax_u8*)(NULL)))) {
        ax_free(((ax_u8*)(self->data)));
        self->data = ((ax_u8*)(NULL));
    }
    self->len = 0;
    self->cap = 0;
}

void ax__AX_std_Vec__u16_push(ax_vec* self, ax_u16 item) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_i64 item_size = ((ax_i64)(sizeof(ax_u16)));
        ax_u16* new_data = ((ax_u16*)(ax_alloc((new_cap * item_size))));
        if ((self->data != ((ax_u16*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * item_size));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    (((ax_u16*)(self->data))[self->len]) = item;
    self->len = (self->len + 1);
}

void ax__AX_std_Vec__u16_destroy(ax_vec* self) {
    if ((self->data != ((ax_u16*)(NULL)))) {
        ax_free(((ax_u8*)(self->data)));
        self->data = ((ax_u16*)(NULL));
    }
    self->len = 0;
    self->cap = 0;
}

ax_vec ax_AX_std_new_vec__string(void) {
    return ((ax_vec){.data=((ax_string*)(NULL)), .len=((ax_i64)(0)), .cap=((ax_i64)(0))});
}

struct ax__AX_std_HashMap__string__string ax_AX_std_new_hashmap__string__string(void) {
    return ((struct ax__AX_std_HashMap__string__string){.keys=((ax_string*)(NULL)), .values=((ax_string*)(NULL)), .hashes=((ax_u64*)(NULL)), .occupied=((ax_bool*)(NULL)), .size=((ax_i64)(0)), .cap=((ax_i64)(0))});
}

void ax__AX_std_HashMap__string__string_insert(struct ax__AX_std_HashMap__string__string* self, ax_string key, ax_string value) {
    if ((self->cap == 0)) {
        self->cap = 16;
        self->keys = ((ax_string*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_string)))))));
        self->values = ((ax_string*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_string)))))));
        self->hashes = ((ax_u64*)(ax_alloc((self->cap * 8))));
        self->occupied = ((ax_bool*)(ax_alloc((self->cap * 1))));
        ax_i64 i = ((ax_i64)(0));
        while ((i < self->cap)) {
            (((ax_bool*)(self->occupied))[i]) = AX_FALSE;
            i = (i + 1);
        }
    }
    if (((self->size * 2) >= self->cap)) {
        ax_i64 old_cap = self->cap;
        ax_string* old_keys = self->keys;
        ax_string* old_values = self->values;
        ax_u64* old_hashes = self->hashes;
        ax_bool* old_occupied = self->occupied;
        self->cap = (old_cap * 2);
        self->keys = ((ax_string*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_string)))))));
        self->values = ((ax_string*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_string)))))));
        self->hashes = ((ax_u64*)(ax_alloc((self->cap * 8))));
        self->occupied = ((ax_bool*)(ax_alloc((self->cap * 1))));
        ax_i64 i = ((ax_i64)(0));
        while ((i < self->cap)) {
            (((ax_bool*)(self->occupied))[i]) = AX_FALSE;
            i = (i + 1);
        }
        self->size = 0;
        ax_i64 j = ((ax_i64)(0));
        while ((j < old_cap)) {
            if ((((ax_bool*)(old_occupied))[j])) {
                ax__AX_std_HashMap__string__string_insert(self, (((ax_string*)(old_keys))[j]), (((ax_string*)(old_values))[j]));
            }
            j = (j + 1);
        }
        if ((old_keys != ((ax_string*)(NULL)))) {
            ax_free(((ax_u8*)(old_keys)));
            ax_free(((ax_u8*)(old_values)));
            ax_free(((ax_u8*)(old_hashes)));
            ax_free(((ax_u8*)(old_occupied)));
        }
    }
    ax_u64 curr_h = ax_AX_std_hash_key__string(key);
    ax_string curr_key = key;
    ax_string curr_value = value;
    ax_i64 idx = ((ax_i64)((curr_h % ((ax_u64)(self->cap)))));
    ax_i64 current_dib = ((ax_i64)(0));
    ax_bool loop = AX_TRUE;
    while (loop) {
        if ((!(((ax_bool*)(self->occupied))[idx]))) {
            (((ax_string*)(self->keys))[idx]) = curr_key;
            (((ax_string*)(self->values))[idx]) = curr_value;
            (((ax_u64*)(self->hashes))[idx]) = curr_h;
            (((ax_bool*)(self->occupied))[idx]) = AX_TRUE;
            self->size = (self->size + 1);
            loop = AX_FALSE;
        } else if ((((((ax_u64*)(self->hashes))[idx]) == curr_h) && ax_str_eq((((ax_string*)(self->keys))[idx]), curr_key))) {
            (((ax_string*)(self->values))[idx]) = curr_value;
            loop = AX_FALSE;
        } else {
            {
                ax_u64 resident_h = (((ax_u64*)(self->hashes))[idx]);
                ax_i64 resident_dib = (((idx - ((ax_i64)((resident_h % ((ax_u64)(self->cap)))))) + self->cap) % self->cap);
                if ((current_dib > resident_dib)) {
                    ax_string tmp_key = (((ax_string*)(self->keys))[idx]);
                    ax_string tmp_value = (((ax_string*)(self->values))[idx]);
                    ax_u64 tmp_h = resident_h;
                    (((ax_string*)(self->keys))[idx]) = curr_key;
                    (((ax_string*)(self->values))[idx]) = curr_value;
                    (((ax_u64*)(self->hashes))[idx]) = curr_h;
                    curr_key = tmp_key;
                    curr_value = tmp_value;
                    curr_h = tmp_h;
                    current_dib = resident_dib;
                }
                idx = ((idx + 1) % self->cap);
                current_dib = (current_dib + 1);
            }
        }
    }
}

static ax_u64 ax_AX_std_hash_key__string(ax_string key) {
    ax_u8* ptr = ((ax_u8*)(&(key)));
    ax_u64 size = sizeof(ax_string);
    if ((size == 16)) {
        ax_string s = (*((ax_string*)(((ax_string*)(ptr)))));
        ax_u64 h = ((ax_u64)(14695981039346656037));
        ax_i64 i = ((ax_i64)(0));
        while ((i < s.len)) {
            h = (h ^ ((ax_u64)((((ax_u8*)(s.ptr))[i]))));
            h = (h * ((ax_u64)(1099511628211)));
            i = (i + 1);
        }
        return h;
    } else {
        {
            ax_u64 h = ((ax_u64)(14695981039346656037));
            ax_i64 i = ((ax_i64)(0));
            while ((i < size)) {
                h = (h ^ ((ax_u64)((((ax_u8*)(ptr))[i]))));
                h = (h * ((ax_u64)(1099511628211)));
                i = (i + 1);
            }
            return h;
        }
    }
}

struct ax__AX_std_Option__string ax__AX_std_Vec__string_get(ax_vec self, ax_i64 index) {
    if (((index < 0) || (index >= self.len))) {
        return ax__AX_std_Option__string_none();
    }
    return ax__AX_std_Option__string_some((((ax_string*)(self.data))[index]));
}

/* Entry point wrapper */
ax_i32 ax_main(void) {
    return ax_main_usr();
}

// Linker stubs for standalone bootstrap tests
ax_i32 ax_ax_driver_load_module(void* mod, struct ax_SymbolTable* st, void* tt) { return 0; }
ax_bool ax_std_string_starts_with(ax_string s, ax_string prefix) { return 0; }
ax_bool ax_std_string_ends_with(ax_string s, ax_string suffix) { return 0; }
ax_bool ax_std_string_contains(ax_string s, ax_string sub) { return 0; }
ax_i64 ax_std_string_char_count(ax_string s) { return 0; }
ax_string ax_std_string_trim(ax_string s) { return s; }
ax_string ax_std_string_to_upper(ax_string s) { return s; }
ax_string ax_std_string_to_lower(ax_string s) { return s; }
