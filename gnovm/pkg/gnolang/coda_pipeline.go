package gnolang

// codaMiddleware is a post-preprocess pass expressed as a single-node
// visitor, matching the TransformB signature. Middleware run within a
// unified TranscribeB walk over the preprocessed tree — the BlockNode
// stack is maintained by the host walk.
//
// Control-flow semantics within a pipeline (see codaPipeline):
//   - TRANS_CONTINUE: let the next middleware run for the same (n, stage);
//     after all middleware, the walk descends into children normally.
//   - TRANS_SKIP: short-circuit the remaining middleware for this node AND
//     skip its children. Used e.g. by the FuncLit guard below.
//   - TRANS_EXIT: abort the entire walk.
//
// A middleware that replaces n (returns a different Node) passes the new
// node to subsequent middleware in the same stage — matching the natural
// chaining behavior of TransformB.
type codaMiddleware func(
	ns []Node, stack []BlockNode, last BlockNode,
	ftype TransField, index int, n Node, stage TransStage,
) (Node, TransCtrl)

// codaPipeline composes middleware into a single TransformB. The middleware
// run in the order given for each (n, stage) pair. Any non-CONTINUE control
// value short-circuits the remaining middleware for that call.
func codaPipeline(mws ...codaMiddleware) TransformB {
	return func(
		ns []Node, stack []BlockNode, last BlockNode,
		ftype TransField, index int, n Node, stage TransStage,
	) (Node, TransCtrl) {
		for _, mw := range mws {
			nn, ctrl := mw(ns, stack, last, ftype, index, n, stage)
			n = nn
			if ctrl != TRANS_CONTINUE {
				return n, ctrl
			}
		}
		return n, TRANS_CONTINUE
	}
}

// skipUnprocessedFuncLitMW short-circuits the pipeline when entering a
// FuncLit that was marked as skipped during preprocess1 (e.g. a FuncLit
// encountered inside a const/type expression whose body was not resolved).
// Descending into such nodes would panic in the coda passes because their
// NameExpr paths and block static data are unresolved. Returning TRANS_SKIP
// skips remaining middleware for this node AND skips its children, which
// is exactly what all coda passes require.
func skipUnprocessedFuncLitMW(
	ns []Node, stack []BlockNode, last BlockNode,
	ftype TransField, index int, n Node, stage TransStage,
) (Node, TransCtrl) {
	if stage == TRANS_ENTER && isSkippedFuncLit(n) {
		return n, TRANS_SKIP
	}
	return n, TRANS_CONTINUE
}
