package gnolang

import "fmt"

type GC struct {
	objs  []*GCObj
	roots []*GCObj
	debug bool
}

type GCObj struct {
	value  TypedValue
	marked bool
	ref    *GCObj
	paths  []*ValuePath
}

func NewGC(debug bool) *GC {
	return &GC{debug: debug}
}

// AddObject use for escaped objects
func (gc *GC) AddObject(obj *GCObj) {
	if gc.debug {
		fmt.Printf("GC: added object: %+v\n", obj)
	}
	gc.objs = append(gc.objs, obj)
}

func (gc *GC) RemoveRoot(path *ValuePath) {
	roots := make([]*GCObj, 0)
	for _, o := range gc.roots {
		var hasPath bool
		for _, valuePath := range o.paths {
			if valuePath == path {
				hasPath = true
				break
			}
		}

		if !hasPath {
			roots = append(roots, o)
			continue
		}
		if gc.debug {
			fmt.Printf("GC: removing root: %+v\n", o)
		}
	}

	gc.roots = roots
}

// AddRoot adds roots that won't be cleaned up by the GC
// use for stack variables/globals
func (gc *GC) AddRoot(root *GCObj) {
	if gc.debug {
		fmt.Printf("GC: add root: %+v\n", root)
	}
	gc.roots = append(gc.roots, root)
}

// when evaluating values that need to escape to the heap
// the VM needs to create a root that hasn't been assigned
// to an identifier yet. so the root it creates has empty path
// this function is to be used at the following operation,
// when evaluating the identifier and setting that path
// to the previously created root with no path
func (gc *GC) setEmptyRootPath(path *ValuePath) {
	root := gc.getRootByPath(nil)
	root.paths = []*ValuePath{path}
	root.ref.paths = []*ValuePath{path}
	if gc.debug {
		fmt.Printf("GC: set root path: %+v\n", root)
	}
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
func (gc *GC) getObjByPath(path *ValuePath) *GCObj {
	for _, obj := range gc.objs {
		for _, valuePath := range obj.paths {
			if valuePath.String() == path.String() {
				return obj
			}
		}
	}
	return nil
}

func (gc *GC) getRootByPath(path *ValuePath) *GCObj {
	for _, obj := range gc.roots {
		if obj.paths == nil && path == nil {
			return obj
		}
		for _, valuePath := range obj.paths {
			if (path == nil && valuePath == nil) || (valuePath.String() == path.String()) {
				return obj
			}
		}
	}
	return nil
}
