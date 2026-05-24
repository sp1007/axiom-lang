import struct
import sys

def parse_obj(filename):
    with open(filename, 'rb') as f:
        data = f.read()
        
    # Read COFF header
    machine, num_sections, timestamp, sym_table_ptr, num_symbols, opt_hdr_size, characteristics = struct.unpack('<HHIIIHH', data[0:20])
    print(f"COFF Header:")
    print(f"  Machine: 0x{machine:X}")
    print(f"  Num Sections: {num_sections}")
    print(f"  Symbol Table Pointer: 0x{sym_table_ptr:X}")
    print(f"  Num Symbols: {num_symbols}")
    
    # Read String Table
    string_table_ptr = sym_table_ptr + num_symbols * 18
    string_table = data[string_table_ptr:]
    
    # Read Symbol Table
    symbols = []
    for i in range(num_symbols):
        off = sym_table_ptr + i * 18
        sym_bytes = data[off:off+18]
        if len(sym_bytes) < 18:
            break
        name_bytes = sym_bytes[0:8]
        if name_bytes[0:4] == b'\x00\x00\x00\x00':
            str_off = struct.unpack('<I', name_bytes[4:8])[0]
            name = string_table[str_off:].split(b'\x00')[0].decode('utf-8', errors='ignore')
        else:
            name = name_bytes.split(b'\x00')[0].decode('utf-8', errors='ignore')
            
        value, sec_num, sym_type, storage_class, num_aux = struct.unpack('<IHHBB', sym_bytes[8:18])
        symbols.append({'name': name, 'value': value, 'sec_num': sec_num, 'type': sym_type, 'storage': storage_class})
        
    print("\nSymbols (Undefined External):")
    for idx, sym in enumerate(symbols):
        if sym['storage'] == 2 and sym['sec_num'] == 0:  # External and Undefined
            print(f"  [{idx}] Name: {sym['name']}")
            
    # Read Section Relocations
    for s_idx in range(num_sections):
        s_off = 20 + s_idx * 40
        sec_name = data[s_off:s_off+8].split(b'\x00')[0].decode('utf-8', errors='ignore')
        
        # Section header structure (40 bytes total):
        # 8 bytes: Name
        # 32 bytes fields:
        vsize, vrva, raw_size, raw_ptr, reloc_ptr, line_ptr, num_relocs, num_lines, characts = struct.unpack('<IIIIIIHHI', data[s_off+8:s_off+40])
        
        if num_relocs > 0:
            print(f"\nRelocations for Section {sec_name} ({num_relocs} relocs):")
            for r_idx in range(num_relocs):
                r_off = reloc_ptr + r_idx * 10
                r_bytes = data[r_off:r_off+10]
                if len(r_bytes) < 10:
                    break
                r_rva, r_sym_idx, r_type = struct.unpack('<IIH', r_bytes)
                sym_name = symbols[r_sym_idx]['name'] if r_sym_idx < len(symbols) else "UNKNOWN"
                print(f"  RVA: 0x{r_rva:X}, SymIdx: {r_sym_idx} ({sym_name}), Type: {r_type}")

if __name__ == '__main__':
    filename = sys.argv[1] if len(sys.argv) > 1 else 'axiom_temp.obj'
    parse_obj(filename)
