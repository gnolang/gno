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
	path   *ValuePath
}

func NewGC(debug bool) *GC {
	return &GC{debug: debug}
}

// AddObject use for escaped objects
func (gc *GC) AddObject(obj *GCObj) {
	if gc.debug {
		fmt.Printf("GC: added object: %p => %+v\n", obj, obj.value)
	}
	gc.objs = append(gc.objs, obj)
}

func (gc *GC) RemoveRoot(path *ValuePath) {
	for i, o := range gc.roots {
		if o.path != path {
			continue
		}

		// we don't care about the order of roots
		// so we can delete it quickly
		gc.roots[i] = gc.roots[len(gc.roots)-1]
		gc.roots = gc.roots[:len(gc.roots)-1]
		if gc.debug {
			fmt.Printf("GC: removing root: VPBlock(%+v,%+v,%+v) => obj: %p\n", o.path.Depth, o.path.Index, o.path.Name, o.ref)
		}
		return
	}
}

// AddRoot adds roots that won't be cleaned up by the GC
// use for stack variables/globals
func (gc *GC) AddRoot(root *GCObj) {
	if gc.debug {
		if root != nil {
			fmt.Printf("GC: add root: %+v => obj: %p\n", root.path, root.ref)
		} else {
			fmt.Printf("GC: add root: <nil> => obj: <nil>\n")
		}
	}
	gc.roots = append(gc.roots, root)
}

// when evaluating values that need to escape to the heap
// the VM needs to create a root that hasn't been assigned
// to an identifier yet. so the root it creates has empty path
// this function is to be used at the following operation,
// when evaluating the identifier and setting that path
// to the previously created root with no path
// it panics if no root is found
func (gc *GC) setEmptyRootPath(path *ValuePath) {
	root := gc.getRootByPath(nil)
	root.path = path
	root.ref.path = path
	if gc.debug {
		fmt.Printf("GC: set root path: %+v\n", root.path)
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
		if obj.path.String() == path.String() {
			return obj
		}
	}
	return nil
}

func (gc *GC) getRootByPath(path *ValuePath) *GCObj {
	for _, obj := range gc.roots {
		if path == nil {
			if obj.path == nil {
				return obj
			}
		} else if obj.path.String() == path.String() {
			return obj
		}
	}
	return nil
}
