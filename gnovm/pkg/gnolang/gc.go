package gnolang

type GC struct {
	objs  []*GCObj
	roots []*GCObj
}

type GCObj struct {
	value  interface{} //todo this should probably be a Pointer struct
	marked bool
	ref    *GCObj
	path   string
}

func NewGC() *GC {
	return &GC{}
}

// AddObject use for escaped objects
func (gc *GC) AddObject(obj *GCObj) {
	gc.objs = append(gc.objs, obj)
}

func (gc *GC) RemoveRoot(path string) {
	for i, o := range gc.roots {
		if o.path != path {
			continue
		}

		gc.roots = append(gc.roots[:i], gc.roots[i+1:]...)

		break
	}
}

// AddRoot adds roots that won't be cleaned up by the GC
// use for stack variables/globals
func (gc *GC) AddRoot(root *GCObj) {
	gc.roots = append(gc.roots, root)
}

// when evaluating values that need to escape to the heap
// the VM needs to create a root that hasn't been assigned
// to an identifier yet. so the root it creates has empty path
// this function is to be used at the following operation,
// when evaluating the identifier and setting that path
// to the previously created root with no path
func (gc *GC) setEmptyRootPath(path string) {
	root := gc.getRootByPath("")
	root.path = path
}

func (gc *GC) Collect() {
	// Mark phase
	for _, root := range gc.roots {
		gc.markObject(root)
	}

	// Sweep phase
	newObjs := make([]*GCObj, 0, len(gc.objs))
	for _, obj := range gc.objs {
		if !obj.marked {
			continue
		}
		obj.marked = false
		newObjs = append(newObjs, obj)
	}
	gc.objs = newObjs
}

func (gc *GC) markObject(obj *GCObj) {
	if obj.marked {
		return
	}

	obj.marked = true

	if obj.ref == nil {
		return
	}
	gc.markObject(obj.ref)
}

// use this only in tests
// because if you hold on to a reference of the GC object
// the Go GC cannot reclaim this memory
// only get GC object references through roots
func (gc *GC) getObjByPath(path string) *GCObj {
	for _, obj := range gc.objs {
		if obj.path == path {
			return obj
		}
	}
	return nil
}

func (gc *GC) getRootByPath(path string) *GCObj {
	for _, obj := range gc.roots {
		if obj.path == path {
			return obj
		}
	}
	return nil
}
