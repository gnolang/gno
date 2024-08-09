package zentasktic

import (
	"testing"

    "gno.land/p/demo/avl"
)

// Shared instance of ZTaskManager
var ztm *ZTaskManager

func init() {
    ztm = NewZTaskManager()
}

func Test_AddTask(t *testing.T) {
    task := Task{Id: "1", RealmId: "1", Body: "First task", ContextId: "1",}

    // Test adding a task successfully.
    err := ztm.AddTask(task)
    if err != nil {
        t.Errorf("Failed to add task: %v", err)
    }

    // Test adding a duplicate task.
    cerr := ztm.AddTask(task)
    if cerr != ErrTaskIdAlreadyExists {
        t.Errorf("Expected ErrTaskIdAlreadyExists, got %v", cerr)
    }
}

func Test_RemoveTask(t *testing.T) {
    
    task := Task{Id: "20", Body: "Removable task", RealmId: "1"}

    // Test adding a task successfully.
    err := ztm.AddTask(task)
    if err != nil {
        t.Errorf("Failed to add task: %v", err)
    }

    retrievedTask, rerr := ztm.GetTaskById(task.Id)
    if rerr != nil {
        t.Errorf("Could not retrieve the added task")
    }

    // Test removing a task
    terr := ztm.RemoveTask(retrievedTask)
    if terr != nil {
        t.Errorf("Expected nil, got %v", terr)
    }
}

func Test_EditTask(t *testing.T) {
    
    task := Task{Id: "2", Body: "First content", RealmId: "1", ContextId: "2"}

    // Test adding a task successfully.
    err := ztm.AddTask(task)
    if err != nil {
        t.Errorf("Failed to add task: %v", err)
    }

    // Test editing the task
    editedTask := Task{Id: task.Id, Body: "Edited content", RealmId: task.RealmId, ContextId: "2"}
    cerr := ztm.EditTask(editedTask)
    if cerr != nil {
        t.Errorf("Failed to edit the task")
    }

    retrievedTask, _ := ztm.GetTaskById(editedTask.Id)
    if retrievedTask.Body != "Edited content" {
        t.Errorf("Task was not edited")
    }
}
/*
func Test_MoveTaskToRealm(t *testing.T) {
    
    task := Task{Id: "3", Body: "First content", RealmId: "1", ContextId: "1"}

    // Test adding a task successfully.
    err := task.AddTask()
    if err != nil {
        t.Errorf("Failed to add task: %v", err)
    }

    // Test moving the task to another realm
    
    cerr := task.MoveTaskToRealm("2")
    if cerr != nil {
        t.Errorf("Failed to move task to another realm")
    }

    retrievedTask, _ := GetTaskById(task.Id)
    if retrievedTask.RealmId != "2" {
        t.Errorf("Task was moved to the wrong realm")
    }
}

func Test_AttachTaskToProject(t *testing.T) {
    
    // Example Projects and Tasks
    prj := Project{Id: "1", Body: "Project 1", RealmId: "1",}
    tsk := Task{Id: "4", Body: "Task 4", RealmId: "1",}

    Projects.Set(prj.Id, prj) // Mock existing project

    tests := []struct {
        name    string
        project Project
        task    Task
        wantErr bool
        errMsg  error
    }{
        {
            name:    "Attach to existing project",
            project: prj,
            task:    tsk,
            wantErr: false,
        },
        {
            name:    "Attach to non-existing project",
            project: Project{Id: "200", Body: "Project 200", RealmId: "1",},
            task:    tsk,
            wantErr: true,
            errMsg:  ErrProjectIdNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.task.AttachTaskToProject(tt.project)
            if (err != nil) != tt.wantErr {
                t.Errorf("AttachTaskToProject() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && err != tt.errMsg {
                t.Errorf("AttachTaskToProject() error = %v, expected %v", err, tt.errMsg)
            }

            // For successful attach, verify the task is added to the project's tasks.
            if !tt.wantErr {
                tasks, exist := ProjectTasks.Get(tt.project.Id)
                if !exist || len(tasks.([]Task)) == 0 {
                    t.Errorf("Task was not attached to the project")
                } else {
                    found := false
                    for _, task := range tasks.([]Task) {
                        if task.Id == tt.task.Id {
                            found = true
                            break
                        }
                    }
                    if !found {
                        t.Errorf("Task was not attached to the project")
                    }
                }
            }
        })
    }
}

func TestDetachTaskFromProject(t *testing.T) {
	
	// Setup:
	project := Project{Id: "p1", Body: "Test Project"}
	task1 := Task{Id: "5", Body: "Task One"}
	task2 := Task{Id: "6", Body: "Task Two"}

	Projects.Set(project.Id, project)
	ProjectTasks.Set(project.Id, []Task{task1, task2})

	tests := []struct {
		name       string
		task       Task
		project    Project
		wantErr    bool
		expectedErr error
	}{
		{
			name:       "Detach existing task from project",
			task:       task1,
			project:    project,
			wantErr:    false,
			expectedErr: nil,
		},
		{
			name:       "Try to detach task from non-existing project",
			task:       task1,
			project:    Project{Id: "nonexistent"},
			wantErr:    true,
			expectedErr: ErrProjectIdNotFound,
		},
		{
			name:       "Try to detach non-existing task from project",
			task:       Task{Id: "nonexistent"},
			project:    project,
			wantErr:    true,
			expectedErr: ErrTaskByIdNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.DetachTaskFromProject(tt.project)

			if tt.wantErr {
				if err == nil || err != tt.expectedErr {
					t.Errorf("%s: expected error %v, got %v", tt.name, tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}

				// For successful detachment, verify the task is no longer part of the project's tasks
				if !tt.wantErr {
					tasks, _ := ProjectTasks.Get(tt.project.Id)
					for _, task := range tasks.([]Task) {
						if task.Id == tt.task.Id {
							t.Errorf("%s: task was not detached from the project", tt.name)
							break
						}
					}
				}
			}
		})
	}
}

func Test_SetTaskDueDate(t *testing.T) {
	taskRealmIdOne, _ := GetTaskById("1")
    taskRealmIdTwo, _ := GetTaskById("10") 
	// Define test cases
	tests := []struct {
		name    string
		task    Task
		dueDate string
		wantErr error
	}{
		{
			name: "Task does not exist",
			task: Task{Id: "nonexistent", RealmId: "2"},
			wantErr: ErrTaskIdNotFound,
		},
		{
			name: "Task not editable due to wrong realm",
			task: taskRealmIdOne,
			wantErr: ErrTaskNotEditable,
		},
		{
			name: "Successfully set due date",
			task: taskRealmIdTwo,
			dueDate: "2023-01-01",
			wantErr: nil,
		},
	}

	// Execute test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.SetTaskDueDate(tc.dueDate)

			// Validate
			if err != tc.wantErr {
				t.Errorf("Expected error %v, got %v", tc.wantErr, err)
			}

			// Additional check for the success case to ensure the due date was actually set
			if err == nil {
				// Fetch the task again to check if the due date was set correctly
				updatedTask, exist := Tasks.Get(tc.task.Id)
				if !exist {
					t.Fatalf("Task %v was not found after setting the due date", tc.task.Id)
				}
				if updatedTask.(Task).Due != tc.dueDate {
					t.Errorf("Expected due date to be %v, got %v", tc.dueDate, updatedTask.(Task).Due)
				}
			}
		})
	}
}

func Test_SetTaskAlert(t *testing.T) {
	taskRealmIdOne, _ := GetTaskById("1")
    taskRealmIdTwo, _ := GetTaskById("10") 
	// Define test cases
	tests := []struct {
		name    string
		task    Task
		alertDate string
		wantErr error
	}{
		{
			name: "Task does not exist",
			task: Task{Id: "nonexistent", RealmId: "2"},
			wantErr: ErrTaskIdNotFound,
		},
		{
			name: "Task not editable due to wrong realm",
			task: taskRealmIdOne,
			wantErr: ErrTaskNotEditable,
		},
		{
			name: "Successfully set alert",
			task: taskRealmIdTwo,
			alertDate: "2024-01-01",
			wantErr: nil,
		},
	}

	// Execute test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.SetTaskAlert(tc.alertDate)

			// Validate
			if err != tc.wantErr {
				t.Errorf("Expected error %v, got %v", tc.wantErr, err)
			}

			// Additional check for the success case to ensure the due date was actually set
			if err == nil {
				// Fetch the task again to check if the due date was set correctly
				updatedTask, exist := Tasks.Get(tc.task.Id)
				if !exist {
					t.Fatalf("Task %v was not found after setting the due date", tc.task.Id)
				}
				if updatedTask.(Task).Alert != tc.alertDate {
					t.Errorf("Expected due date to be %v, got %v", tc.alertDate, updatedTask.(Task).Due)
				}
			}
		})
	}
}

// getters

func Test_GetAllTasks(t *testing.T) {
    
    // mocking the tasks based on previous tests
    // TODO: add isolation?
    knownTasks := []Task{
        {Id: "1", RealmId: "1", Body: "First task", ContextId: "1",},
        {Id: "10", Body: "First content", RealmId: "2", ContextId: "2", Due: "2023-01-01", Alert: "2024-01-01"},
        {Id: "2", Body: "Edited content", RealmId: "1", ContextId: "2",},
        {Id: "20", Body: "Removable task", RealmId: "1",},
        {Id: "3", Body: "First content", RealmId: "2", ContextId: "1",},
        {Id: "40", Body: "Task 40", RealmId: "1",},
    }

    // Manually marshal the known tasks to create the expected outcome.
    tasksObject := TasksObject{Tasks: knownTasks}
    expected, err := tasksObject.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal known tasks: %v", err)
    }

    // Execute GetAllTasks() to get the actual outcome.
    actual, err := GetAllTasks()
    if err != nil {
        t.Fatalf("GetAllTasks() failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual task JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}

func Test_GetTasksByDate(t *testing.T) {
	
	tests := []struct {
		name      string
		taskDate  string
		filterType string
		want      string
		wantErr   bool
	}{
		{"SpecificDate", "2023-01-01", "specific", `{"tasks":[{"taskId":"10","taskProjectId":"","taskContextId":"2","taskRealmId":"2","taskBody":"First content","taskDue":"2023-01-01","taskAlert":"2024-01-01"}]}`, false},
		{"BeforeDate", "2022-04-05", "before", "", false},
		{"AfterDate", "2023-04-05", "after", "", false},
		{"NoMatch", "2002-04-07", "specific", "", false},
		{"InvalidDateFormat", "April 5, 2023", "specific", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTasksByDate(tt.taskDate, tt.filterType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTasksByDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("GetTasksByDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetTaskById(t *testing.T){
    // test getting a non-existing task
    nonTask, err := GetTaskById("0")
    if err != ErrTaskByIdNotFound {
        t.Fatalf("Expected ErrTaskByIdNotFound, got: %v", err)
    }

    // test getting the correct task by id
    correctTask, err := GetTaskById("1")
    if err != nil {
        t.Fatalf("Failed to get task by id, error: %v", err)
    }

    if correctTask.Body != "First task" {
        t.Fatalf("Got the wrong task, with body: %v", correctTask.Body)
    }
}

func Test_GetTasksByRealm(t *testing.T) {
    
    // mocking the tasks based on previous tests
    // TODO: add isolation?
    tasksInAssessRealm := []Task{
        {Id: "1", RealmId: "1", Body: "First task", ContextId: "1",},
        {Id: "2", RealmId: "1", Body: "Edited content", ContextId: "2",},
        {Id: "20", Body: "Removable task", RealmId: "1",},
        {Id: "40", Body: "Task 40", RealmId: "1",},
    }

    // Manually marshal the known tasks to create the expected outcome.
    tasksObjectAssess := TasksObject{Tasks: tasksInAssessRealm}
    expected, err := tasksObjectAssess.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal tasks in Assess: %v", err)
    }

    actual, err := GetTasksByRealm("1")
    if err != nil {
        t.Fatalf("GetTasksByRealm('1') failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual task JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}

func Test_GetTasksByContext(t *testing.T) {
    
    // mocking the tasks based on previous tests
    // TODO: add isolation?
    tasksInContextOne := []Task{
        {Id: "1", RealmId: "1", Body: "First task", ContextId: "1",},
        {Id: "3", RealmId: "2", Body: "First content", ContextId: "1",},
    }

    // Manually marshal the known tasks to create the expected outcome.
    tasksObjectForContexts := TasksObject{Tasks: tasksInContextOne}
    expected, err := tasksObjectForContexts.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal tasks for ContextId 1: %v", err)
    }

    actual, err := GetTasksByContext("1")
    if err != nil {
        t.Fatalf("GetTasksByContext('1') failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual task JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}
*/
