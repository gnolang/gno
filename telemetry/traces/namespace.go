package traces

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type namespace string

type namespaceContext struct {
	namespace namespace
	ctx       context.Context
}

const (
	NamespaceVMInit     namespace = "vmInit"
	NamespaceVMProcess  namespace = "vmProcess"
	NamespaceVMQuery    namespace = "vmQuery"
	NamespaceMachineRun namespace = "machineRun"
)

// Maps goroutine number to namespace and context.
var namespaces = make(map[int]namespaceContext)

func InitNamespace(ctx context.Context, ns namespace) {
	if ctx == nil {
		ctx = context.Background()
	}

	namespaces[goroutineID()] = namespaceContext{namespace: ns, ctx: ctx}
}

func goroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func ActiveNamespace() namespace {
	id := goroutineID()
	namespaceCtx, ok := namespaces[id]
	if !ok {
		panic("should not happen.")
	}

	return namespaceCtx.namespace
}
