package gnolang

type TransformB func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl)

// TranscribeB handles the common routine of keeping a stack and popping it at
// leave. Unlike Transcribe, this function will pop the blocknode from the
// stack even if a skip is called within, therefore it is safer to use.
//
// Transform node `n` of `ftype`/`index` in context `ns` during `stage` with
// parent block node `last`.  Return a new node to replace the old one, or the
// node will be deleted (or set to nil).  NOTE: Do not mutate stack or ns.
//
// Returns:
//   - TRANS_CONTINUE to visit children recursively;
//   - TRANS_SKIP to skip the following stages for the node
//     (BLOCK/BLOCK2/LEAVE), but a skip from LEAVE will
//     skip the following stages of the parent.
//   - TRANS_EXIT to stop traversing altogether.
//
// XXX Replace all usage of Transcribe() with TranscribeB().
func TranscribeB(last BlockNode, n Node, tb TransformB) (nn Node) {
	// create stack of BlockNodes.
	stack := append(make([]BlockNode, 0, 32), last)

	// Iterate over all nodes recursively.
	nn = Transcribe(n, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (nn Node, tctrl TransCtrl) {
		switch stage {
		case TRANS_ENTER:
			nn, tctrl = tb(ns, stack, last, ftype, index, n, stage) // user transform
			return
		case TRANS_BLOCK:
			pushInitBlock(n.(BlockNode), &last, &stack)             // push block
			nn, tctrl = tb(ns, stack, last, ftype, index, n, stage) // user transform
			if tctrl == TRANS_SKIP {                                // if skip,
				switch n.(type) { /*                         */ //   pop block
				case BlockNode:
					stack = stack[:len(stack)-1]
					last = stack[len(stack)-1]
				}
			}
			return
		case TRANS_BLOCK2:
			nn, tctrl = tb(ns, stack, last, ftype, index, n, stage) // user transform
			if tctrl == TRANS_SKIP {                                // if skip,
				switch n.(type) { /*                         */ //   pop block
				case BlockNode:
					stack = stack[:len(stack)-1]
					last = stack[len(stack)-1]
				}
			}
			return
		case TRANS_LEAVE:
			nn, tctrl = tb(ns, stack, last, ftype, index, n, stage) // user transform
			switch n.(type) {                                       // pop block
			case BlockNode:
				stack = stack[:len(stack)-1]
				last = stack[len(stack)-1]
			}
			return
		default:
			panic("unexpected transcribe stage")
		}
	})

	return nn
}
