package gnolang

type GC struct {
	objs  []*GCObj
	roots []*GCObj
}

type GCObj struct {
	key    MapKey
	marked bool
	refs   []*GCObj
}

func (o *GCObj) AddRef(obj *GCObj) {
	o.refs = append(o.refs, obj)
}

func NewGC() *GC {
	return &GC{}
}

// AddObject use for escaped objects
func (gc *GC) AddObject(obj *GCObj) {
	gc.objs = append(gc.objs, obj)
}

func (gc *GC) RemoveObject(key MapKey) {
	for i, o := range gc.objs {
		if o.key == key {
			gc.objs = append(gc.objs[:i], gc.objs[i+1:]...)
			break
		}
	}
}

// AddRoot adds roots that won't be cleaned up by the GC
// use for stack variables/globals
func (gc *GC) AddRoot(root *GCObj) {
	gc.roots = append(gc.roots, root)
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
	for _, edge := range obj.refs {
		gc.markObject(edge)
	}
}

func (gc *GC) getObj(id MapKey) *GCObj {
	for _, obj := range gc.objs {
		if obj.key == id {
			return obj
		}
	}
	return nil
}
