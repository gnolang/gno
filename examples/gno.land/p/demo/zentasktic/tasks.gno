package zentasktic

import (
	"time"

	"gno.land/p/demo/avl"
)

type Task struct {
	Id 			string `json:"taskId"`
	ProjectId 	string `json:"taskProjectId"`
	ContextId	string `json:"taskContextId"`
	RealmId 	string `json:"taskRealmId"`
	Body 		string `json:"taskBody"`
	Due			string `json:"taskDue"`
	Alert		string `json:"taskAlert"`
}

type ZTaskManager struct {
	Tasks *avl.Tree
}

func NewZTaskManager() *ZTaskManager {
	return &ZTaskManager{
		Tasks: avl.NewTree(),
	}
}

// actions

func (ztm *ZTaskManager) AddTask(t Task) error {
	if ztm.Tasks.Size() != 0 {
		_, exist := ztm.Tasks.Get(t.Id)
		if exist {
			return ErrTaskIdAlreadyExists
		}
	}
	ztm.Tasks.Set(t.Id, t)
	return nil
}

func (ztm *ZTaskManager) RemoveTask(t Task) error {
	existingTaskInterface, exist := ztm.Tasks.Get(t.Id)
	if !exist {
		return ErrTaskIdNotFound
	}
	existingTask := existingTaskInterface.(Task)

	 // task is removable only in Asses (RealmId 1) or via a Collection (RealmId 4)
	if existingTask.RealmId != "1" && existingTask.RealmId != "4" {
		return ErrTaskNotRemovable
	}

	_, removed := ztm.Tasks.Remove(existingTask.Id)
	if !removed {
		return ErrTaskNotRemoved
	}
	return nil
}

func (ztm *ZTaskManager) EditTask(t Task) error {
	existingTaskInterface, exist := ztm.Tasks.Get(t.Id)
	if !exist {
		return ErrTaskIdNotFound
	}
	existingTask := existingTaskInterface.(Task)

	// task Body is editable only when task is in Assess, RealmId = "1"
	if t.RealmId != "1" {
		if t.Body != existingTask.Body {
			return ErrTaskNotInAssessRealm
		}
	}

	ztm.Tasks.Set(t.Id, t)
	return nil
}

// Helper function to move a task to a different realm
func (ztm *ZTaskManager) MoveTaskToRealm(taskId, realmId string) error {
	existingTaskInterface, exist := ztm.Tasks.Get(taskId)
	if !exist {
		return ErrTaskIdNotFound
	}
	existingTask := existingTaskInterface.(Task)
	existingTask.RealmId = realmId
	ztm.Tasks.Set(taskId, existingTask)
	return nil
}

func (ztm *ZTaskManager) SetTaskDueDate(taskId, dueDate string) error {
	taskInterface, exist := ztm.Tasks.Get(taskId)
	if !exist {
		return ErrTaskIdNotFound
	}
	task := taskInterface.(Task)

	if task.RealmId == "2" {
		task.Due = dueDate
		ztm.Tasks.Set(task.Id, task)
	} else {
		return ErrTaskNotEditable
	}

	return nil
}

func (ztm *ZTaskManager) SetTaskAlert(taskId, alertDate string) error {
	taskInterface, exist := ztm.Tasks.Get(taskId)
	if !exist {
		return ErrTaskIdNotFound
	}
	task := taskInterface.(Task)

	if task.RealmId == "2" {
		task.Alert = alertDate
		ztm.Tasks.Set(task.Id, task)
	} else {
		return ErrTaskNotEditable
	}

	return nil
}

// tasks & projects association

func (zpm *ZProjectManager) AttachTaskToProject(ztm *ZTaskManager, t Task, p Project) error {
	existingProjectInterface, exist := zpm.Projects.Get(p.Id)
	if !exist {
		return ErrProjectIdNotFound
	}
	existingProject := existingProjectInterface.(Project)

	existingProjectTasksInterface, texist := zpm.ProjectTasks.Get(p.Id)
	if !texist {
		existingProject.Tasks = []Task{}
	} else {
		tasks, ok := existingProjectTasksInterface.([]Task)
		if !ok {
			return ErrProjectTasksNotFound
		}
		existingProject.Tasks = tasks
	}

	t.ProjectId = p.Id
	// @todo we need to remove it from Tasks if it was previously added there, then detached
	existingTask, err := ztm.GetTaskById(t.Id)
	if err == nil {
		ztm.RemoveTask(existingTask)
	}
	updatedTasks := append(existingProject.Tasks, t)
	zpm.ProjectTasks.Set(p.Id, updatedTasks)

	return nil
}

func (zpm *ZProjectManager) EditProjectTask(projectTaskId string, projectTaskBody string, projectId string) error {
    existingProjectInterface, exist := zpm.Projects.Get(projectId)
    if !exist {
        return ErrProjectIdNotFound
    }
    existingProject := existingProjectInterface.(Project)

	if existingProject.RealmId != "1" {
		return ErrProjectNotEditable
	}

    existingProjectTasksInterface, texist := zpm.ProjectTasks.Get(projectId)
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
        existingProject.Tasks[index].Body = projectTaskBody
    } else {
        return ErrTaskByIdNotFound
    }

    zpm.ProjectTasks.Set(projectId, existingProject.Tasks)
    return nil
}

func (zpm *ZProjectManager) DetachTaskFromProject(ztm *ZTaskManager, projectTaskId string, detachedTaskId string, p Project) error {
	existingProjectInterface, exist := zpm.Projects.Get(p.Id)
	if !exist {
		return ErrProjectIdNotFound
	}
	existingProject := existingProjectInterface.(Project)

	existingProjectTasksInterface, texist := zpm.ProjectTasks.Get(p.Id)
	if !texist {
		return ErrProjectTasksNotFound
	}
	tasks, ok := existingProjectTasksInterface.([]Task)
	if !ok {
		return ErrProjectTasksNotFound
	}
	existingProject.Tasks = tasks

	var foundTask Task
	var index int = -1
	for i, task := range existingProject.Tasks {
		if task.Id == projectTaskId {
			index = i
			foundTask = task
			break
		}
	}

	if index != -1 {
		existingProject.Tasks = append(existingProject.Tasks[:index], existingProject.Tasks[index+1:]...)
	} else {
		return ErrTaskByIdNotFound
	}

	foundTask.ProjectId = ""
	foundTask.Id = detachedTaskId
	// Tasks and ProjectTasks have different storage, if a task is detached from a Project
	// we add it to the Tasks storage
	if err := ztm.AddTask(foundTask); err != nil {
		return err
	}

	zpm.ProjectTasks.Set(p.Id, existingProject.Tasks)
	return nil
}


func (zpm *ZProjectManager) RemoveTaskFromProject(projectTaskId string, projectId string) error {
	existingProjectInterface, exist := zpm.Projects.Get(projectId)
	if !exist {
		return ErrProjectIdNotFound
	}
	existingProject := existingProjectInterface.(Project)

	existingProjectTasksInterface, texist := zpm.ProjectTasks.Get(projectId)
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
		existingProject.Tasks = append(existingProject.Tasks[:index], existingProject.Tasks[index+1:]...)
	} else {
		return ErrTaskByIdNotFound
	}

	zpm.ProjectTasks.Set(projectId, existingProject.Tasks)
	return nil
}

// getters

func (ztm *ZTaskManager) GetTaskById(taskId string) (Task, error) {
	if ztm.Tasks.Size() != 0 {
		tInterface, exist := ztm.Tasks.Get(taskId)
		if exist {
			return tInterface.(Task), nil
		}
	}
	return Task{}, ErrTaskIdNotFound
}

func (ztm *ZTaskManager) GetAllTasks() (task string) {
	var allTasks []Task

	ztm.Tasks.Iterate("", "", func(key string, value interface{}) bool {
		if task, ok := value.(Task); ok {
			allTasks = append(allTasks, task)
		}
		return false
	})
	// Create a TasksObject with all collected tasks.
	tasksObject := &TasksObject{
		Tasks: allTasks,
	}

	// Use the custom MarshalJSON method to marshal the tasks into JSON.
	marshalledTasks, merr := tasksObject.MarshalJSON()
	if merr != nil {
		return ""
	} 
	return string(marshalledTasks)	
}

func (ztm *ZTaskManager) GetTasksByRealm(realmId string) (tasks string) {
	var realmTasks []Task

	ztm.Tasks.Iterate("", "", func(key string, value interface{}) bool {
		if task, ok := value.(Task); ok {
			if task.RealmId == realmId {
				realmTasks = append(realmTasks, task)
			}
		}
		return false
	})

	// Create a TasksObject with all collected tasks.
	tasksObject := &TasksObject{
		Tasks: realmTasks,
	}

	// Use the custom MarshalJSON method to marshal the tasks into JSON.
	marshalledTasks, merr := tasksObject.MarshalJSON()
	if merr != nil {
			return ""
	} 
	return string(marshalledTasks)
}

func (ztm *ZTaskManager) GetTasksByContextAndRealm(contextId string, realmId string) (tasks string) {
	var contextTasks []Task

	ztm.Tasks.Iterate("", "", func(key string, value interface{}) bool {
		if task, ok := value.(Task); ok {
			if task.ContextId == contextId && task.ContextId == realmId {
				contextTasks = append(contextTasks, task)
			}
		}
		return false
	})

	// Create a TasksObject with all collected tasks.
	tasksObject := &TasksObject{
		Tasks: contextTasks,
	}

	// Use the custom MarshalJSON method to marshal the tasks into JSON.
	marshalledTasks, merr := tasksObject.MarshalJSON()
	if merr != nil {
		return ""
	} 
	return string(marshalledTasks)
}

func (ztm *ZTaskManager) GetTasksByDate(taskDate string, filterType string) (tasks string) {
	parsedDate, err := time.Parse("2006-01-02", taskDate)
	if err != nil {
		return ""
	}

	var filteredTasks []Task

	ztm.Tasks.Iterate("", "", func(key string, value interface{}) bool {
		task, ok := value.(Task)
		if !ok {
			return false
		}

		storedDate, serr := time.Parse("2006-01-02", task.Due)
		if serr != nil {
			return false
		}

		switch filterType {
		case "specific":
			if storedDate.Format("2006-01-02") == parsedDate.Format("2006-01-02") {
				filteredTasks = append(filteredTasks, task)
			}
		case "before":
			if storedDate.Before(parsedDate) {
				filteredTasks = append(filteredTasks, task)
			}
		case "after":
			if storedDate.After(parsedDate) {
				filteredTasks = append(filteredTasks, task)
			}
		}

		return false
	})

	// Create a TasksObject with all collected tasks.
	tasksObject := &TasksObject{
		Tasks: filteredTasks,
	}

	// Use the custom MarshalJSON method to marshal the tasks into JSON.
	marshalledTasks, merr := tasksObject.MarshalJSON()
	if merr != nil {
		return ""
	} 
	return string(marshalledTasks)
}
