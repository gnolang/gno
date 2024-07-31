// base implementation
package zentasktic

import (
	"gno.land/p/demo/avl"
)


type Collection struct {
	Id 			string `json:"collectionId"`
	RealmId 	string `json:"collectionRealmId"`
	Name 		string `json:"collectionName"`
	Tasks		[]Task `json:"collectionTasks"`
	Projects	[]Project `json:"collectionProjects"`
}

type ZCollectionManager struct {
	Collections *avl.Tree 
	CollectionTasks *avl.Tree
	CollectionProjects *avl.Tree 
}

func NewZCollectionManager() *ZCollectionManager {
    return &ZCollectionManager{
        Collections: avl.NewTree(),
        CollectionTasks: avl.NewTree(),
        CollectionProjects: avl.NewTree(),
    }
}


// actions

func (zcolm *ZCollectionManager) AddCollection(c Collection) (err error) {
	// implementation
	if zcolm.Collections.Size() != 0 {
		_, exist := zcolm.Collections.Get(c.Id)
		if exist {
			return ErrCollectionIdAlreadyExists
		}
	}
	zcolm.Collections.Set(c.Id, c)
	return nil
}

func (zcolm *ZCollectionManager) EditCollection(c Collection) (err error) {
	// implementation
	if zcolm.Collections.Size() != 0 {
		_, exist := zcolm.Collections.Get(c.Id)
		if !exist {
			return ErrCollectionIdNotFound
		}
	}
	
	zcolm.Collections.Set(c.Id, c)
	return nil
}

func (zcolm *ZCollectionManager) RemoveCollection(c Collection) (err error) {
    // implementation
    if zcolm.Collections.Size() != 0 {
        collectionInterface, exist := zcolm.Collections.Get(c.Id)
        if !exist {
            return ErrCollectionIdNotFound
        }
        collection := collectionInterface.(Collection)

        _, removed := zcolm.Collections.Remove(collection.Id)
        if !removed {
            return ErrCollectionNotRemoved
        }

        if zcolm.CollectionTasks.Size() != 0 {
            _, removedTasks := zcolm.CollectionTasks.Remove(collection.Id)
            if !removedTasks {
                return ErrCollectionNotRemoved
            }	
        }

        if zcolm.CollectionProjects.Size() != 0 {
            _, removedProjects := zcolm.CollectionProjects.Remove(collection.Id)
            if !removedProjects {
                return ErrCollectionNotRemoved
            }	
        }
    }
    return nil
}


func (zcolm *ZCollectionManager) AddProjectToCollection(zpm *ZProjectManager, c Collection, p Project) (err error) {
	// implementation
	if zcolm.Collections.Size() != 0 {
		_, exist := zcolm.Collections.Get(c.Id)
		if !exist {
			return ErrCollectionIdNotFound
		}
	}

	existingCollectionProjects, texist := zcolm.CollectionProjects.Get(c.Id)
	if !texist {
		// If the collections has no projects yet, initialize the slice.
		existingCollectionProjects = []Project{}
	} else {
		projects, ok := existingCollectionProjects.([]Project)
		if !ok {
			return ErrCollectionsProjectsNotFound
		}
		existingCollectionProjects = projects
	}
	p.RealmId = "4"
	if err := zpm.EditProject(p); err != nil {
		return err
	}
	updatedProjects := append(existingCollectionProjects.([]Project), p)
	zcolm.CollectionProjects.Set(c.Id, updatedProjects)

	return nil
}

func (zcolm *ZCollectionManager) AddTaskToCollection(ztm *ZTaskManager, c Collection, t Task) (err error) {
	// implementation
	if zcolm.Collections.Size() != 0 {
		_, exist := zcolm.Collections.Get(c.Id)
		if !exist {
			return ErrCollectionIdNotFound
		}
	}

	existingCollectionTasks, texist := zcolm.CollectionTasks.Get(c.Id)
	if !texist {
		// If the collections has no tasks yet, initialize the slice.
		existingCollectionTasks = []Task{}
	} else {
		tasks, ok := existingCollectionTasks.([]Task)
		if !ok {
			return ErrCollectionsTasksNotFound
		}
		existingCollectionTasks = tasks
	}
	t.RealmId = "4"
	if err := ztm.EditTask(t); err != nil {
		return err
	}
	updatedTasks := append(existingCollectionTasks.([]Task), t)
	zcolm.CollectionTasks.Set(c.Id, updatedTasks)

	return nil
}

func (zcolm *ZCollectionManager) RemoveProjectFromCollection(zpm *ZProjectManager, c Collection, p Project) (err error) {
	// implementation
	if zcolm.Collections.Size() != 0 {
		_, exist := zcolm.Collections.Get(c.Id)
		if !exist {
			return ErrCollectionIdNotFound
		}
	}

	existingCollectionProjects, texist := zcolm.CollectionProjects.Get(c.Id)
	if !texist {
		// If the collection has no projects yet, return appropriate error
		return ErrCollectionsProjectsNotFound
	}

	// Find the index of the project to be removed.
	var index int = -1
	for i, project := range existingCollectionProjects.([]Project) {
		if project.Id == p.Id {
			index = i
			break
		}
	}

	// If the project was found, we remove it from the slice.
	if index != -1 {
		// by default we send it back to Assess
		p.RealmId = "1"
		zpm.EditProject(p)
		existingCollectionProjects = append(existingCollectionProjects.([]Project)[:index], existingCollectionProjects.([]Project)[index+1:]...)
	} else {
		// Project not found in the collection
		return ErrProjectByIdNotFound 
	}
	zcolm.CollectionProjects.Set(c.Id, existingCollectionProjects)

	return nil
}

func (zcolm *ZCollectionManager) RemoveTaskFromCollection(ztm *ZTaskManager, c Collection, t Task) (err error) {
	// implementation
	if zcolm.Collections.Size() != 0 {
		_, exist := zcolm.Collections.Get(c.Id)
		if !exist {
			return ErrCollectionIdNotFound
		}
	}

	existingCollectionTasks, texist := zcolm.CollectionTasks.Get(c.Id)
	if !texist {
		// If the collection has no tasks yet, return appropriate error
		return ErrCollectionsTasksNotFound
	}

	// Find the index of the task to be removed.
	var index int = -1
	for i, task := range existingCollectionTasks.([]Task) {
		if task.Id == t.Id {
			index = i
			break
		}
	}

	// If the task was found, we remove it from the slice.
	if index != -1 {
		// by default, we send the task to Assess
		t.RealmId = "1"
		ztm.EditTask(t)
		existingCollectionTasks = append(existingCollectionTasks.([]Task)[:index], existingCollectionTasks.([]Task)[index+1:]...)
	} else {
		// Task not found in the collection
		return ErrTaskByIdNotFound 
	}
	zcolm.CollectionTasks.Set(c.Id, existingCollectionTasks)

	return nil
}

// getters

func (zcolm *ZCollectionManager) GetCollectionById(collectionId string) (Collection, error) {
    if zcolm.Collections.Size() != 0 {
        cInterface, exist := zcolm.Collections.Get(collectionId)
        if exist {
            collection := cInterface.(Collection)
            // look for collection Tasks, Projects
            existingCollectionTasks, texist := zcolm.CollectionTasks.Get(collectionId)
            if texist {
                collection.Tasks = existingCollectionTasks.([]Task)
            }
            existingCollectionProjects, pexist := zcolm.CollectionProjects.Get(collectionId)
            if pexist {
                collection.Projects = existingCollectionProjects.([]Project)
            }
            return collection, nil
        }
        return Collection{}, ErrCollectionByIdNotFound
    }
    return Collection{}, ErrCollectionByIdNotFound
}

func (zcolm *ZCollectionManager) GetCollectionTasks(c Collection) (tasks []Task, err error) {
	
	if zcolm.CollectionTasks.Size() != 0 {
		task, exist := zcolm.CollectionTasks.Get(c.Id)
		if !exist {
			// if there's no record in CollectionTasks, we don't have to return anything
			return nil, ErrCollectionsTasksNotFound
		} else {
			// type assertion to convert interface{} to []Task
			existingCollectionTasks, ok := task.([]Task)
			if !ok {
				return nil, ErrTaskFailedToAssert
			}
			return existingCollectionTasks, nil
		}
	}
	return nil, nil
}

func (zcolm *ZCollectionManager) GetCollectionProjects(c Collection) (projects []Project, err error) {
	
	if zcolm.CollectionProjects.Size() != 0 {
		project, exist := zcolm.CollectionProjects.Get(c.Id)
		if !exist {
			// if there's no record in CollectionProjets, we don't have to return anything
			return nil, ErrCollectionsProjectsNotFound
		} else {
			// type assertion to convert interface{} to []Projet
			existingCollectionProjects, ok := project.([]Project)
			if !ok {
				return nil, ErrProjectFailedToAssert
			}
			return existingCollectionProjects, nil
		}
	}
	return nil, nil
}

func (zcolm *ZCollectionManager) GetAllCollections() (collections string, err error) {
	// implementation
	var allCollections []Collection
	
	// Iterate over the Collections AVL tree to collect all Project objects.
	
	zcolm.Collections.Iterate("", "", func(key string, value interface{}) bool {
		if collection, ok := value.(Collection); ok {
			// get collection tasks, if any
			collectionTasks, _ := zcolm.GetCollectionTasks(collection)
			if collectionTasks != nil {
				collection.Tasks = collectionTasks
			}
			// get collection prokects, if any
			collectionProjects, _ := zcolm.GetCollectionProjects(collection)
			if collectionProjects != nil {
				collection.Projects = collectionProjects
			}
			allCollections = append(allCollections, collection)
		}
		return false // Continue iteration until all nodes have been visited.
	})

	// Create a CollectionsObject with all collected tasks.
	collectionsObject := CollectionsObject{
		Collections: allCollections,
	}

	// Use the custom MarshalJSON method to marshal the collections into JSON.
	marshalledCollections, merr := collectionsObject.MarshalJSON()
	if merr != nil {
		return "", merr
	} 
	return string(marshalledCollections), nil
} 