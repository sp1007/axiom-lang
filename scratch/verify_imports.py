import sys
import ctypes
import os

imports = {
    "kernel32.dll": [
        "VirtualAlloc", "VirtualFree", "ExitProcess", "GetFileAttributesA", 
        "CreateDirectoryA", "DeleteFileA", "RemoveDirectoryA", "MoveFileA", 
        "CopyFileA", "GetCommandLineA", "GetEnvironmentVariableA", "GetCommandLineW"
    ],
    "ax_runtime.dll": [
        "ax_actor_step", "ax_actor_is_running", "ax_actor_has_messages"
    ],
    "ucrtbase.dll": [
        "memset", "abort", "memcpy", "puts", "fflush", "strlen", "printf", 
        "sprintf", "atof", "exit", "fopen", "fputs", "fprintf", "fclose", 
        "snprintf", "fwrite", "fseek", "ftell", "rewind", "fread", "system"
    ]
}

if hasattr(os, 'add_dll_directory'):
    os.add_dll_directory(r"d:\projects\compiler\Axiom")
    os.add_dll_directory(r"d:\projects\compiler\Axiom\bin")

for dll_name, symbols in imports.items():
    print(f"\nChecking {dll_name}...")
    try:
        if dll_name == "ax_runtime.dll":
            dll = ctypes.CDLL(os.path.join(r"d:\projects\compiler\Axiom", dll_name))
        else:
            dll = ctypes.CDLL(dll_name)
    except Exception as e:
        print(f"  FAILED to load {dll_name}: {e}")
        continue
        
    for sym in symbols:
        try:
            addr = getattr(dll, sym)
            # print(f"  {sym}: OK")
        except AttributeError:
            print(f"  {sym}: MISSING!")
