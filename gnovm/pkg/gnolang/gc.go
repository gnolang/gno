package gnolang

type GC struct {
	objs  []*GCObj
	roots []*GCObj
}

type GCObj struct {
	id     ObjectID
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

func (gc *GC) RemoveObject(id ObjectID) {
	for i, o := range gc.objs {
		if o.id == id {
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

func (gc *GC) Collect() []ObjectID {
	// Mark phase
	for _, root := range gc.roots {
		gc.markObject(root)
	}

	// Sweep phase
	var newObjs []*GCObj
	var deletedIDs []ObjectID
	for _, obj := range gc.objs {
		if !obj.marked {
			deletedIDs = append(deletedIDs, obj.id)
			continue
		}
		obj.marked = false
		newObjs = append(newObjs, obj)
	}
	gc.objs = newObjs
	return deletedIDs
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

func (gc *GC) getObj(id ObjectID) *GCObj {
	for _, obj := range gc.objs {
		if obj.id == id {
			return obj
		}
	}
	return nil
}
