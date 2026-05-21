package testing

func Test_B_RunParallel_Unsupported(t *T) {
	expectPanic(t, "testing: B.RunParallel is unsupported in Gno", func() {
		b := &B{}
		b.RunParallel(nil)
	})
}

func Test_B_SetParallelism_Unsupported(t *T) {
	expectPanic(t, "testing: B.SetParallelism is unsupported in Gno", func() {
		b := &B{}
		b.SetParallelism(1)
	})
}

func Test_B_Nested_Run_Panics(t *T) {
	expectPanic(t, "testing: nested B.Run is not yet supported", func() {
		parent := &B{}
		sub := &B{parent: parent}
		sub.Run("inner", func(b *B) {})
	})
}

func Test_B_Run_PausesAndResumesParentTimer(t *T) {
	on := &B{N: 1, name: "bench"}
	on.StartTimer()
	on.Run("sub", func(sub *B) {})
	if !on.timerOn {
		t.Errorf("parent timer should resume to ON after b.Run returned")
	}
	on.StopTimer()

	off := &B{N: 1, name: "bench2"}
	off.Run("sub", func(sub *B) {})
	if off.timerOn {
		t.Errorf("parent timer should stay OFF when it wasn't running before b.Run")
	}
}

func Test_B_Setenv_Unsupported(t *T) {
	expectPanic(t, "testing: B.Setenv is not supported", func() {
		b := &B{}
		b.Setenv("FOO", "BAR")
	})
}

func Test_B_Cleanup_Unsupported(t *T) {
	expectPanic(t, "testing: B.Cleanup is not supported", func() {
		b := &B{}
		b.Cleanup(func() {})
	})
}

func Test_B_ReportMetric_Unsupported(t *T) {
	expectPanic(t, "testing: B.ReportMetric is not supported", func() {
		b := &B{}
		b.ReportMetric(1.0, "x")
	})
}

func Test_B_TempDir_Unsupported(t *T) {
	expectPanic(t, "testing: B.TempDir is not supported", func() {
		b := &B{}
		b.TempDir()
	})
}

func Test_PB_Next_Unsupported(t *T) {
	expectPanic(t, "testing: PB.Next is unsupported in Gno", func() {
		pb := &PB{}
		pb.Next()
	})
}

func expectPanic(t *T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic %q, got none", want)
			return
		}
		msg, ok := r.(string)
		if !ok {
			t.Errorf("expected string panic %q, got %T: %v", want, r, r)
			return
		}
		if msg != want {
			t.Errorf("panic message mismatch:\n  got:  %q\n  want: %q", msg, want)
		}
	}()
	fn()
}
