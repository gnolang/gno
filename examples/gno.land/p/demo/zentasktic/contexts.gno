// base implementation
package zentasktic

import (
	"gno.land/p/demo/avl"
)

type Context struct {
	Id   string `json:"contextId"`
	Name string `json:"contextName"`
}

type ZContextManager struct {
	Contexts *avl.Tree
}

func NewZContextManager() *ZContextManager {
	return &ZContextManager{
		Contexts: avl.NewTree(),
	}
}

// Actions

func (zcm *ZContextManager) AddContext(c Context) error {
	if zcm.Contexts.Size() != 0 {
		_, exist := zcm.Contexts.Get(c.Id)
		if exist {
			return ErrContextIdAlreadyExists
		}
	}
	zcm.Contexts.Set(c.Id, c)
	return nil
}

func (zcm *ZContextManager) EditContext(c Context) error {
	if zcm.Contexts.Size() != 0 {
		_, exist := zcm.Contexts.Get(c.Id)
		if !exist {
			return ErrContextIdNotFound
		}
	}
	zcm.Contexts.Set(c.Id, c)
	return nil
}

func (zcm *ZContextManager) RemoveContext(c Context) error {
	if zcm.Contexts.Size() != 0 {
		context, exist := zcm.Contexts.Get(c.Id)
		if !exist {
			return ErrContextIdNotFound
		}
		_, removed := zcm.Contexts.Remove(context.(Context).Id)
		if !removed {
			return ErrContextNotRemoved
		}
	}
	return nil
}

func (zcm *ZContextManager) AddContextToTask(ztm *ZTaskManager, c Context, t Task) error {
	taskInterface, exist := ztm.Tasks.Get(t.Id)
	if !exist {
		return ErrTaskIdNotFound
	}
	_, cexist := zcm.Contexts.Get(c.Id)
	if !cexist {
		return ErrContextIdNotFound
	}

	if t.RealmId == "2" {
		task := taskInterface.(Task)
		task.ContextId = c.Id
		ztm.Tasks.Set(t.Id, task)
	} else {
		return ErrTaskNotEditable
	}

	return nil
}

func (zcm *ZContextManager) AddContextToProject(zpm *ZProjectManager, c Context, p Project) error {
	projectInterface, exist := zpm.Projects.Get(p.Id)
	if !exist {
		return ErrProjectIdNotFound
	}
	_, cexist := zcm.Contexts.Get(c.Id)
	if !cexist {
		return ErrContextIdNotFound
	}

	if p.RealmId == "2" {
		project := projectInterface.(Project)
		project.ContextId = c.Id
		zpm.Projects.Set(p.Id, project)
	} else {
		return ErrProjectNotEditable
	}

	return nil
}

func (zcm *ZContextManager) AddContextToProjectTask(zpm *ZProjectManager, c Context, p Project, projectTaskId string) error {
	
	_, cexist := zcm.Contexts.Get(c.Id)
	if !cexist {
		return ErrContextIdNotFound
	}

	existingProjectInterface, exist := zpm.Projects.Get(p.Id)
    if !exist {
        return ErrProjectIdNotFound
    }
    existingProject := existingProjectInterface.(Project)

	if existingProject.RealmId != "2" {
		return ErrProjectNotEditable
	}

    existingProjectTasksInterface, texist := zpm.ProjectTasks.Get(p.Id)
    if !texist {
        return ErrProjectTasksNotFound
    }
    tasks, ok := existingProjectTasksInterface.([]Task)
    if !ok {
        return ErrProjectTasksNotFound
    }
    existingProject.Tasks = tasks

    var index int = -1
    for i, task := range existingProject.Tasks {
        if task.Id == projectTaskId {
            index = i
            break
        }
    }

    if index != -1 {
        existingProject.Tasks[index].ContextId = c.Id
    } else {
        return ErrTaskByIdNotFound
    }

    zpm.ProjectTasks.Set(p.Id, existingProject.Tasks)
    return nil
}

// getters

func (zcm *ZContextManager) GetContextById(contextId string) (Context, error) {
	if zcm.Contexts.Size() != 0 {
		cInterface, exist := zcm.Contexts.Get(contextId)
		if exist {
			return cInterface.(Context), nil
		}
		return Context{}, ErrContextIdNotFound
	}
	return Context{}, ErrContextIdNotFound
}

func (zcm *ZContextManager) GetAllContexts() (string) {
	var allContexts []Context

	// Iterate over the Contexts AVL tree to collect all Context objects.
	zcm.Contexts.Iterate("", "", func(key string, value interface{}) bool {
		if context, ok := value.(Context); ok {
			allContexts = append(allContexts, context)
		}
		return false // Continue iteration until all nodes have been visited.
	})

	// Create a ContextsObject with all collected contexts.
	contextsObject := &ContextsObject{
		Contexts: allContexts,
	}

	// Use the custom MarshalJSON method to marshal the contexts into JSON.
	marshalledContexts, merr := contextsObject.MarshalJSON()
	if merr != nil {
		return ""
	}
	return string(marshalledContexts)
}

