package gnolang

import (
	"tinygo.org/x/go-llvm"
)

func compileFactoriel() (*llvm.ExecutionEngine, llvm.Value) {
	llvm.LinkInMCJIT()
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()

	mod := llvm.NewModule("fac_module")

	fac_args := []llvm.Type{llvm.Int32Type()}
	fac_type := llvm.FunctionType(llvm.Int32Type(), fac_args, false)
	fac := llvm.AddFunction(mod, "fac", fac_type)
	fac.SetFunctionCallConv(llvm.CCallConv)
	n := fac.Param(0)

	entry := llvm.AddBasicBlock(fac, "entry")
	iftrue := llvm.AddBasicBlock(fac, "iftrue")
	iffalse := llvm.AddBasicBlock(fac, "iffalse")
	end := llvm.AddBasicBlock(fac, "end")

	builder := llvm.NewBuilder()
	defer builder.Dispose()

	builder.SetInsertPointAtEnd(entry)
	If := builder.CreateICmp(llvm.IntEQ, n, llvm.ConstInt(llvm.Int32Type(), 0, false), "cmptmp")
	builder.CreateCondBr(If, iftrue, iffalse)

	builder.SetInsertPointAtEnd(iftrue)
	res_iftrue := llvm.ConstInt(llvm.Int32Type(), 1, false)
	builder.CreateBr(end)

	builder.SetInsertPointAtEnd(iffalse)
	n_minus := builder.CreateSub(n, llvm.ConstInt(llvm.Int32Type(), 1, false), "subtmp")
	call_fac_args := []llvm.Value{n_minus}
	call_fac := builder.CreateCall(fac_type, fac, call_fac_args, "calltmp")
	res_iffalse := builder.CreateMul(n, call_fac, "multmp")
	builder.CreateBr(end)

	builder.SetInsertPointAtEnd(end)
	res := builder.CreatePHI(llvm.Int32Type(), "result")
	phi_vals := []llvm.Value{res_iftrue, res_iffalse}
	phi_blocks := []llvm.BasicBlock{iftrue, iffalse}
	res.AddIncoming(phi_vals, phi_blocks)
	builder.CreateRet(res)

	err := llvm.VerifyModule(mod, llvm.ReturnStatusAction)
	if err != nil {
		panic(err)
	}

	options := llvm.NewMCJITCompilerOptions()
	options.SetMCJITOptimizationLevel(3)
	options.SetMCJITEnableFastISel(true)
	options.SetMCJITNoFramePointerElim(true)
	options.SetMCJITCodeModel(llvm.CodeModelJITDefault)
	engine, err := llvm.NewMCJITCompiler(mod, options)
	if err != nil {
		panic(err)
	}

	pass := llvm.NewPassManager()
	defer pass.Dispose()

	pass.AddSCCPPass()
	pass.AddInstructionCombiningPass()
	pass.AddPromoteMemoryToRegisterPass()
	pass.AddGVNPass()
	pass.AddCFGSimplificationPass()
	pass.Run(mod)

	return &engine, fac
}

//opt levl is 0 to 3
/*
	-O0 (default): This is the default optimization level, and it means that no optimization is performed. This can be useful for debugging, since it preserves the original code structure and makes it easier to step through the code.

	-O1: This optimization level performs some simple optimizations that don't take much time to run. This includes things like inlining small functions, simplifying expressions, and removing dead code.

	-O2: This optimization level performs more aggressive optimizations that take more time to run. This includes things like loop unrolling, function inlining, and instruction scheduling.

	-O3: This optimization level performs even more aggressive optimizations than -O2. This can result in faster code, but can also increase compilation time and code size. This includes things like interprocedural optimization, loop vectorization, and function specialization.
*/
