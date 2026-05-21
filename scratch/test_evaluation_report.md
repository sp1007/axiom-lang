# AXIOM Compiler Test Suite Evaluation Report

This report summarizes the compilation results of all `*.ax` files found in the `tests/` directory and its subdirectories.

## Summary

- **Total Files Scanned:** 58
- **Successfully Compiled:** 8
- **Failed to Compile:** 50

### Breakdown by Category

- **air**: 2/7 passed
- **root**: 0/19 passed
- **generics**: 4/15 passed
- **lexer**: 0/2 passed
- **sema**: 2/15 passed

---

## Detailed Results

| File Path | Category | Status | Expected Behavior / Details |
|---|---|---|---|
| [001_return_const.ax](file:///d:\projects\compiler\Axiom\tests\air\001_return_const.ax) | air | ✅ Pass | Should Compile |
| [002_var_decl.ax](file:///d:\projects\compiler\Axiom\tests\air\002_var_decl.ax) | air | ✅ Pass | Should Compile |
| [003_if_stmt.ax](file:///d:\projects\compiler\Axiom\tests\air\003_if_stmt.ax) | air | ❌ Fail | Should Compile (Error: `[31merror[E2010][0m: [1mundefined: 'x'[0m`) |
| [004_if_else.ax](file:///d:\projects\compiler\Axiom\tests\air\004_if_else.ax) | air | ❌ Fail | Should Compile (Error: `[31merror[E2010][0m: [1mundefined: 'x'[0m`) |
| [005_while_loop.ax](file:///d:\projects\compiler\Axiom\tests\air\005_while_loop.ax) | air | ❌ Fail | Should Compile (Error: `[31merror[E2010][0m: [1mundefined: 'x'[0m`) |
| [006_multi_func.ax](file:///d:\projects\compiler\Axiom\tests\air\006_multi_func.ax) | air | ❌ Fail | Should Compile (Error: `[31merror[E3005][0m: [1mreturn type mismatch: expected 3, found void[0m`) |
| [007_bool_lit.ax](file:///d:\projects\compiler\Axiom\tests\air\007_bool_lit.ax) | air | ❌ Fail | Should Compile (Error: `[31merror[E3005][0m: [1mreturn type mismatch: expected 3, found void[0m`) |
| [axiom_ai_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_ai_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_compiler_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_compiler_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E1000][0m: [1mexpected declaration, got identifier[0m`) |
| [axiom_compliance_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_compliance_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0002][0m: [1munterminated string literal[0m`) |
| [axiom_dbms_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_dbms_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E1000][0m: [1mexpected expression, got ','[0m`) |
| [axiom_devops_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_devops_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E1000][0m: [1mexpected expression, got 'fn'[0m`) |
| [axiom_distributed_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_distributed_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E1000][0m: [1mexpected newline, got 'as'[0m`) |
| [axiom_dsl_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_dsl_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '#' (0x23)[0m`) |
| [axiom_ffi_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_ffi_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_functional_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_functional_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_game_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_game_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_gui_complex_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_gui_complex_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '$' (0x24)[0m`) |
| [axiom_gui_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_gui_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_hft_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_hft_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E1000][0m: [1mexpected newline, got 'as'[0m`) |
| [axiom_lowlevel_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_lowlevel_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_multimedia_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_multimedia_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_robotics_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_robotics_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [axiom_scientific_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_scientific_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '#' (0x23)[0m`) |
| [axiom_security_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_security_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E1000][0m: [1mexpected identifier, got '('[0m`) |
| [axiom_web_suite.ax](file:///d:\projects\compiler\Axiom\tests\axiom_web_suite.ax) | root | ❌ Fail | Future Language Spec (Not yet supported by Stage 1 Parser) (Error: `[31merror[E0004][0m: [1munexpected character '@' (0x40)[0m`) |
| [async_await_outside.ax](file:///d:\projects\compiler\Axiom\tests\generics\async_await_outside.ax) | generics | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3010][0m: [1mawait requires Future[T], found 22[0m`) |
| [async_basic.ax](file:///d:\projects\compiler\Axiom\tests\generics\async_basic.ax) | generics | ❌ Fail | Should Compile (Error: `[31merror[E3010][0m: [1mawait requires Future[T], found 22[0m`) |
| [async_nested.ax](file:///d:\projects\compiler\Axiom\tests\generics\async_nested.ax) | generics | ❌ Fail | Should Compile (Error: `[31merror[E3010][0m: [1mawait requires Future[T], found 22[0m`) |
| [generic_constrained.ax](file:///d:\projects\compiler\Axiom\tests\generics\generic_constrained.ax) | generics | ❌ Fail | Should Compile (Error: `panic: CTypeName: unresolved generic type parameter (TypeID 22)`) |
| [generic_identity.ax](file:///d:\projects\compiler\Axiom\tests\generics\generic_identity.ax) | generics | ❌ Fail | Should Compile (Error: `[31merror[E3001][0m: [1mtype mismatch: expected 22, found 3[0m`) |
| [generic_pair.ax](file:///d:\projects\compiler\Axiom\tests\generics\generic_pair.ax) | generics | ❌ Fail | Should Compile (Error: `[31merror[E3001][0m: [1mtype mismatch: expected 24, found 3[0m`) |
| [generic_stack.ax](file:///d:\projects\compiler\Axiom\tests\generics\generic_stack.ax) | generics | ❌ Fail | Should Compile (Error: `[31merror[E3001][0m: [1mtype mismatch: expected 23, found 3[0m`) |
| [interface_basic.ax](file:///d:\projects\compiler\Axiom\tests\generics\interface_basic.ax) | generics | ✅ Pass | Should Compile |
| [interface_missing.ax](file:///d:\projects\compiler\Axiom\tests\generics\interface_missing.ax) | generics | ✅ Pass | Expected Semantic/Syntax Error |
| [sum_type_color.ax](file:///d:\projects\compiler\Axiom\tests\generics\sum_type_color.ax) | generics | ❌ Fail | Should Compile (Error: `axc: C compilation failed: C compiler (gcc) failed:`) |
| [sum_type_nonexhaustive.ax](file:///d:\projects\compiler\Axiom\tests\generics\sum_type_nonexhaustive.ax) | generics | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3005][0m: [1mreturn type mismatch: expected 10, found 0[0m`) |
| [sum_type_result.ax](file:///d:\projects\compiler\Axiom\tests\generics\sum_type_result.ax) | generics | ❌ Fail | Should Compile (Error: `axc: C compilation failed: C compiler (gcc) failed:`) |
| [valid_const_decl.ax](file:///d:\projects\compiler\Axiom\tests\generics\valid_const_decl.ax) | generics | ✅ Pass | Should Compile |
| [valid_struct_methods.ax](file:///d:\projects\compiler\Axiom\tests\generics\valid_struct_methods.ax) | generics | ✅ Pass | Should Compile |
| [valid_sum_type_option.ax](file:///d:\projects\compiler\Axiom\tests\generics\valid_sum_type_option.ax) | generics | ❌ Fail | Should Compile (Error: `axc: C compilation failed: C compiler (gcc) failed:`) |
| [hello.ax](file:///d:\projects\compiler\Axiom\tests\lexer\hello.ax) | lexer | ❌ Fail | Should Compile (Error: `[31merror[E2010][0m: [1mundefined: 'println'[0m`) |
| [test_hello.ax](file:///d:\projects\compiler\Axiom\tests\lexer\test_hello.ax) | lexer | ❌ Fail | Should Compile (Error: `[31merror[E2005][0m: [1mfield not found in module[0m`) |
| [err_arg_type.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_arg_type.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3001][0m: [1mtype mismatch: expected 3, found 12[0m`) |
| [err_assign_mismatch.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_assign_mismatch.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E1000][0m: [1mexpected identifier, got 'mut'[0m`) |
| [err_bad_condition.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_bad_condition.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3011][0m: [1mif condition must be bool, found 3[0m`) |
| [err_call_args.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_call_args.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3005][0m: [1mreturn type mismatch: expected 14, found 3[0m`) |
| [err_call_non_func.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_call_non_func.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `axc: C compilation failed: C compiler (gcc) failed:`) |
| [err_immutable.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_immutable.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3010][0m: [1mcannot assign to immutable variable 'x'[0m`) |
| [err_redefine.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_redefine.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E2001][0m: [1msymbol already defined in this scope: 'x'[0m`) |
| [err_return_type.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_return_type.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3005][0m: [1mreturn type mismatch: expected 3, found 12[0m`) |
| [err_type_mismatch.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_type_mismatch.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E3001][0m: [1mtype mismatch: expected 3, found 12[0m`) |
| [err_undefined.ax](file:///d:\projects\compiler\Axiom\tests\sema\err_undefined.ax) | sema | ❌ Fail | Expected Semantic/Syntax Error (Error: `[31merror[E2010][0m: [1mundefined: 'y'[0m`) |
| [valid_assign.ax](file:///d:\projects\compiler\Axiom\tests\sema\valid_assign.ax) | sema | ❌ Fail | Should Compile (Error: `axc: C compilation failed: C compiler (gcc) failed:`) |
| [valid_fibonacci.ax](file:///d:\projects\compiler\Axiom\tests\sema\valid_fibonacci.ax) | sema | ✅ Pass | Should Compile |
| [valid_hello.ax](file:///d:\projects\compiler\Axiom\tests\sema\valid_hello.ax) | sema | ❌ Fail | Should Compile (Error: `[31merror[E1000][0m: [1mexpected identifier, got '.'[0m`) |
| [valid_shadow.ax](file:///d:\projects\compiler\Axiom\tests\sema\valid_shadow.ax) | sema | ✅ Pass | Should Compile |
| [valid_uninit.ax](file:///d:\projects\compiler\Axiom\tests\sema\valid_uninit.ax) | sema | ❌ Fail | Should Compile (Error: `axc: C compilation failed: C compiler (gcc) failed:`) |
