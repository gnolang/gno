package runner

type TestResult struct {
	PanicOccurred bool        // panic 여부
	PanicMessage  interface{} // panic 메시지
	Error         error       // 반환된 에러
	Result        interface{} // 반환된 값
}

func monitorRuning(fn interface{}, args ...interface{}) TestResult
