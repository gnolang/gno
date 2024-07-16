package zentasktic_core

import (
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/zentasktic"
)

type Workable interface {
	// restrict implementation of Workable to this realm
	assertWorkable()
}

type isWorkable struct {}

func (wt *WorkableTask) assertWorkable() {}

func (wp *WorkableProject) assertWorkable() {}

var _ Workable = &WorkableTask{}
var _ Workable = &WorkableProject{}

type WorkDuration struct {
	Id				string `json:"workDurationId"`
    ObjectId 		string `json:"objectId"`
    ObjectType      string `json:"objectType"` // Task, Project
	Duration  		time.Duration `json:"workDuration"`
	Workable   		Workable      `json:"-"`
}

var (
	WorkDurations avl.Tree // wd.Id -> wd
)

func (wd WorkDuration) AddWorkDuration() (err error) { 
	if WorkDurations.Size() != 0 {
		_, exist := WorkDurations.Get(wd.Id)
		if exist {
			return ErrWorkDurationIdAlreadyExists
		}
	}
	WorkDurations.Set(wd.Id, wd)
	return nil
}

func (wd WorkDuration) EditWorkDuration() (err error) { 
	
	if WorkDurations.Size() != 0 {
		_, exist := WorkDurations.Get(wd.Id)
		if !exist {
			return ErrWorkDurationNotFound
		}
	}
	WorkDurations.Set(wd.Id, wd)
	return nil
}

func (wd WorkDuration) RemoveWorkDuration() (err error) { 
	// implementation
	if WorkDurations.Size() != 0 {
		_, exist := WorkDurations.Get(wd.Id)
		if !exist {
			return ErrWorkDurationNotFound
		}
	}
	
	_, removed := WorkDurations.Remove(wd.Id)
	if !removed {
		return ErrWorkDurationNotRemoved
	}
	return nil
}

// getters

func GetWorkDurationByOjectId(objectId string, objectType string) (wd WorkDuration, err error) {
	// implementation
	workDuration := WorkDuration{}

	// Iterate over the WorkDuration AVL tree to see if we have a matching ObjectId.
	
	WorkDurations.Iterate("", "", func(key string, value interface{}) bool {
		if workDurationItem, ok := value.(WorkDuration); ok {
			if workDurationItem.ObjectId == objectId && workDurationItem.ObjectType == objectType {
				workDuration = workDurationItem
			}
		}
		return false // Continue iteration until all nodes have been visited.
	})

	return workDuration, nil
}

func GetAllWorkDurations() (workDurations []WorkDuration, err error){
	// implementation
	var allWorkDurations []WorkDuration

	// Iterate over the WorkDurations AVL tree to collect all WorkDurations objects.
	if WorkDurations.Size() == 0 {
		return nil, ErrWorkDurationsEmpty
	} else {
		WorkDurations.Iterate("", "", func(key string, value interface{}) bool {
			if workDuration, ok := value.(WorkDuration); ok {
				allWorkDurations = append(allWorkDurations, workDuration)
			}
			return false // Continue iteration until all nodes have been visited.
		})
	}

	return allWorkDurations, nil
}