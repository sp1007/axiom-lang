import sys
import struct

def parse_pe_imports(filename):
    with open(filename, 'rb') as f:
        data = f.read()
    
    # Verify MZ signature
    if data[0:2] != b'MZ':
        print("Not a valid MZ/PE file")
        return
        
    pe_offset = struct.unpack('<I', data[0x3C:0x40])[0]
    if data[pe_offset:pe_offset+4] != b'PE\x00\x00':
        print("Not a valid PE file")
        return
        
    magic = struct.unpack('<H', data[pe_offset+24:pe_offset+26])[0]
    is_pe32_plus = magic == 0x20b
    
    # Optional header directories
    dir_offset = pe_offset + 24 + (112 if is_pe32_plus else 96)
    import_rva, import_size = struct.unpack('<II', data[dir_offset+8:dir_offset+16])
    
    if import_size == 0:
        print("No imports found")
        return
        
    # Find section headers to map RVA to file offset
    num_sections = struct.unpack('<H', data[pe_offset+6:pe_offset+8])[0]
    section_offset = pe_offset + 24 + struct.unpack('<H', data[pe_offset+20:pe_offset+22])[0]
    
    sections = []
    for i in range(num_sections):
        off = section_offset + i * 40
        name = data[off:off+8].rstrip(b'\x00').decode('utf-8', errors='ignore')
        vsize, vrva, raw_size, raw_ptr = struct.unpack('<IIII', data[off+8:off+24])
        sections.append({'name': name, 'vrva': vrva, 'vsize': vsize, 'raw_ptr': raw_ptr, 'raw_size': raw_size})
        
    def rva_to_offset(rva):
        for sec in sections:
            if sec['vrva'] <= rva < sec['vrva'] + sec['vsize']:
                return sec['raw_ptr'] + (rva - sec['vrva'])
        return None
        
    import_offset = rva_to_offset(import_rva)
    if import_offset is None:
        print(f"Cannot map import RVA 0x{import_rva:X} to file offset")
        return
        
    # Read Import Directory Table
    idx = import_offset
    while True:
        entry = data[idx:idx+20]
        if len(entry) < 20 or entry == b'\x00'*20:
            break
            
        original_first_thunk, timestamp, forwarder_chain, name_rva, first_thunk = struct.unpack('<IIIII', entry)
        idx += 20
        
        dll_name_offset = rva_to_offset(name_rva)
        if dll_name_offset is None:
            continue
        dll_name = data[dll_name_offset:].split(b'\x00')[0].decode('utf-8', errors='ignore')
        print(f"\nDLL: {dll_name}")
        
        # Read thunks
        thunk_rva = original_first_thunk if original_first_thunk != 0 else first_thunk
        thunk_offset = rva_to_offset(thunk_rva)
        if thunk_offset is None:
            continue
            
        t_idx = thunk_offset
        step = 8 if is_pe32_plus else 4
        fmt = '<Q' if is_pe32_plus else '<I'
        
        while True:
            val_bytes = data[t_idx:t_idx+step]
            if len(val_bytes) < step:
                break
            val = struct.unpack(fmt, val_bytes)[0]
            if val == 0:
                break
            t_idx += step
            
            # Check if import by ordinal
            is_ordinal = (val & (1 << (63 if is_pe32_plus else 31))) != 0
            if is_ordinal:
                ordinal = val & 0xFFFF
                print(f"  Ordinal: {ordinal}")
            else:
                name_offset = rva_to_offset(val & 0x7FFFFFFF)
                if name_offset is not None:
                    hint = struct.unpack('<H', data[name_offset:name_offset+2])[0]
                    imported_name = data[name_offset+2:].split(b'\x00')[0].decode('utf-8', errors='ignore')
                    print(f"  {imported_name} (hint: {hint})")

if __name__ == '__main__':
    filename = sys.argv[1] if len(sys.argv) > 1 else 'minimal.exe'
    parse_pe_imports(filename)
