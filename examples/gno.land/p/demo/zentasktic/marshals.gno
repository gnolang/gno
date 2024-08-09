package zentasktic

import (
	"bytes"
)


type ContextsObject struct {
	Contexts	[]Context
}

type TasksObject struct {
	Tasks	[]Task
}

type ProjectsObject struct {
	Projects	[]Project
}

type CollectionsObject struct {
	Collections	[]Collection
}

type RealmsObject struct {
	Realms	[]Realm
}

type ObjectJourney struct {
	ObjectPaths []ObjectPath
}

func (c Context) MarshalJSON() ([]byte, error) {

	var b bytes.Buffer
	
	b.WriteByte('{')

	b.WriteString(`"contextId":"`)
	b.WriteString(c.Id)
	b.WriteString(`",`)

	b.WriteString(`"contextName":"`)
	b.WriteString(c.Name)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (cs ContextsObject) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"contexts":[`)
	
	for i, context := range cs.Contexts {
		if i > 0 {
			b.WriteString(`,`)
		}
		contextJSON, cerr := context.MarshalJSON()
		if cerr == nil {
			b.WriteString(string(contextJSON))
		}
	}

	b.WriteString(`]}`)

	return b.Bytes(), nil
}

func (t Task) MarshalJSON() ([]byte, error) {

	var b bytes.Buffer
	
	b.WriteByte('{')

	b.WriteString(`"taskId":"`)
	b.WriteString(t.Id)
	b.WriteString(`",`)

	b.WriteString(`"taskProjectId":"`)
	b.WriteString(t.ProjectId)
	b.WriteString(`",`)

	b.WriteString(`"taskContextId":"`)
	b.WriteString(t.ContextId)
	b.WriteString(`",`)

	b.WriteString(`"taskRealmId":"`)
	b.WriteString(t.RealmId)
	b.WriteString(`",`)

	b.WriteString(`"taskBody":"`)
	b.WriteString(t.Body)
	b.WriteString(`",`)

	b.WriteString(`"taskDue":"`)
	b.WriteString(t.Due)
	b.WriteString(`",`)

	b.WriteString(`"taskAlert":"`)
	b.WriteString(t.Alert)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (ts TasksObject) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"tasks":[`)
	for i, task := range ts.Tasks {
		if i > 0 {
			b.WriteString(`,`)
		}
		taskJSON, cerr := task.MarshalJSON()
		if cerr == nil {
			b.WriteString(string(taskJSON))
		}
	}
	b.WriteString(`]}`)

	return b.Bytes(), nil
}

func (p Project) MarshalJSON() ([]byte, error) {

	var b bytes.Buffer
	
	b.WriteByte('{')

	b.WriteString(`"projectId":"`)
	b.WriteString(p.Id)
	b.WriteString(`",`)

	b.WriteString(`"projectContextId":"`)
	b.WriteString(p.ContextId)
	b.WriteString(`",`)

	b.WriteString(`"projectRealmId":"`)
	b.WriteString(p.RealmId)
	b.WriteString(`",`)

	b.WriteString(`"projectTasks":[`)

	if len(p.Tasks) != 0 {
		for i, projectTask := range p.Tasks {
			if i > 0 {
				b.WriteString(`,`)
			}
			projectTaskJSON, perr := projectTask.MarshalJSON()
			if perr == nil {
				b.WriteString(string(projectTaskJSON))
			}
		}
	}

	b.WriteString(`],`)

	b.WriteString(`"projectBody":"`)
	b.WriteString(p.Body)
	b.WriteString(`",`)

	b.WriteString(`"projectDue":"`)
	b.WriteString(p.Due)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (ps ProjectsObject) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"projects":[`)
	for i, project := range ps.Projects {
		if i > 0 {
			b.WriteString(`,`)
		}
		projectJSON, perr := project.MarshalJSON()
		if perr == nil {
			b.WriteString(string(projectJSON))
		}
	}
	b.WriteString(`]}`)

	return b.Bytes(), nil
}

func (c Collection) MarshalJSON() ([]byte, error) {

	var b bytes.Buffer
	
	b.WriteByte('{')

	b.WriteString(`"collectionId":"`)
	b.WriteString(c.Id)
	b.WriteString(`",`)

	b.WriteString(`"collectionRealmId":"`)
	b.WriteString(c.RealmId)
	b.WriteString(`",`)

	b.WriteString(`"collectionName":"`)
	b.WriteString(c.Name)
	b.WriteString(`",`)

	b.WriteString(`"collectionTasks":[`)
	for i, collectionTask := range c.Tasks {
		if i > 0 {
			b.WriteString(`,`)
		}
		collectionTaskJSON, perr := collectionTask.MarshalJSON()
		if perr == nil {
			b.WriteString(string(collectionTaskJSON))
		}
	}
	b.WriteString(`],`)

	b.WriteString(`"collectionProjects":[`)
	for i, collectionProject := range c.Projects {
		if i > 0 {
			b.WriteString(`,`)
		}
		collectionProjectJSON, perr := collectionProject.MarshalJSON()
		if perr == nil {
			b.WriteString(string(collectionProjectJSON))
		}
	}
	b.WriteString(`],`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (co CollectionsObject) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"collections":[`)
	for i, collection := range co.Collections {
		if i > 0 {
			b.WriteString(`,`)
		}
		collectionJSON, perr := collection.MarshalJSON()
		if perr == nil {
			b.WriteString(string(collectionJSON))
		}
	}
	b.WriteString(`]}`)

	return b.Bytes(), nil
}

func (r Realm) MarshalJSON() ([]byte, error) {

	var b bytes.Buffer
	
	b.WriteByte('{')

	b.WriteString(`"realmId":"`)
	b.WriteString(r.Id)
	b.WriteString(`",`)

	b.WriteString(`"realmName":"`)
	b.WriteString(r.Name)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (rs RealmsObject) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"realms":[`)
	
	for i, realm := range rs.Realms {
		if i > 0 {
			b.WriteString(`,`)
		}
		realmJSON, rerr := realm.MarshalJSON()
		if rerr == nil {
			b.WriteString(string(realmJSON))
		}
	}

	b.WriteString(`]}`)

	return b.Bytes(), nil
}

func (op ObjectPath) MarshalJSON() ([]byte, error) {

	var b bytes.Buffer
	
	b.WriteByte('{')

	b.WriteString(`"objectType":"`)
	b.WriteString(op.ObjectType)
	b.WriteString(`",`)

	b.WriteString(`"id":"`)
	b.WriteString(op.Id)
	b.WriteString(`",`)

	b.WriteString(`"realmId":"`)
	b.WriteString(op.RealmId)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (oj ObjectJourney) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"objectJourney":[`)
	
	for i, objectPath := range oj.ObjectPaths {
		if i > 0 {
			b.WriteString(`,`)
		}
		objectPathJSON, oerr := objectPath.MarshalJSON()
		if oerr == nil {
			b.WriteString(string(objectPathJSON))
		}
	}

	b.WriteString(`]}`)

	return b.Bytes(), nil
}
