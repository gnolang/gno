package zentasktic

import (
	"testing"

    "gno.land/p/demo/avl"
)
/*
func Test_AddProject(t *testing.T) {
    
    project := Project{Id: "1", RealmId: "1", Body: "First project", ContextId: "1",}

    // Test adding a project successfully.
    err := project.AddProject()
    if err != nil {
        t.Errorf("Failed to add project: %v", err)
    }

    // Test adding a duplicate project.
    cerr := project.AddProject()
    if cerr != ErrProjectIdAlreadyExists {
        t.Errorf("Expected ErrProjectIdAlreadyExists, got %v", cerr)
    }
}


func Test_RemoveProject(t *testing.T) {
    
    project := Project{Id: "20", Body: "Removable project", RealmId: "1", ContextId: "2",}

    // Test adding a project successfully.
    err := project.AddProject()
    if err != nil {
        t.Errorf("Failed to add project: %v", err)
    }

    retrievedProject, rerr := GetProjectById(project.Id)
    if rerr != nil {
        t.Errorf("Could not retrieve the added project")
    }

    // Test removing a project
    terr := retrievedProject.RemoveProject()
    if terr != ErrProjectNotRemoved {
        t.Errorf("Expected ErrProjectNotRemoved, got %v", terr)
    }
}


func Test_EditProject(t *testing.T) {
    
    project := Project{Id: "2", Body: "Second project content", RealmId: "1", ContextId: "2",}

    // Test adding a project successfully.
    err := project.AddProject()
    if err != nil {
        t.Errorf("Failed to add project: %v", err)
    }

    // Test editing the project
    editedProject := Project{Id: project.Id, Body: "Edited project content", RealmId: project.RealmId, ContextId: "2",}
    cerr := editedProject.EditProject()
    if cerr != nil {
        t.Errorf("Failed to edit the project")
    }

    retrievedProject, _ := GetProjectById(editedProject.Id)
    if retrievedProject.Body != "Edited project content" {
        t.Errorf("Project was not edited")
    }
}


func Test_MoveProjectToRealm(t *testing.T) {
    
    project := Project{Id: "3", Body: "Project id 3 content", RealmId: "1", ContextId: "1",}

    // Test adding a project successfully.
    err := project.AddProject()
    if err != nil {
        t.Errorf("Failed to add project: %v", err)
    }

    // Test moving the project to another realm
    
    cerr := project.MoveProjectToRealm("2")
    if cerr != nil {
        t.Errorf("Failed to move project to another realm")
    }

    retrievedProject, _ := GetProjectById(project.Id)
    if retrievedProject.RealmId != "2" {
        t.Errorf("Project was moved to the wrong realm")
    }
}

func Test_SetProjectDueDate(t *testing.T) {
	projectRealmIdOne, _ := GetProjectById("1")
    projectRealmIdTwo, _ := GetProjectById("10") 
	// Define test cases
	tests := []struct {
		name    string
		project    Project
		dueDate string
		wantErr error
	}{
		{
			name: "Project does not exist",
			project: Project{Id: "nonexistent", RealmId: "2"},
			wantErr: ErrProjectIdNotFound,
		},
		{
			name: "Project not editable due to wrong realm",
			project: projectRealmIdOne,
			wantErr: ErrProjectNotEditable,
		},
		{
			name: "Successfully set alert",
			project: projectRealmIdTwo,
			dueDate: "2024-01-01",
			wantErr: nil,
		},
	}

	// Execute test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.project.SetProjectDueDate(tc.dueDate)

			// Validate
			if err != tc.wantErr {
				t.Errorf("Expected error %v, got %v", tc.wantErr, err)
			}

			// Additional check for the success case to ensure the due date was actually set
			if err == nil {
				// Fetch the task again to check if the due date was set correctly
				updatedProject, exist := Projects.Get(tc.project.Id)
				if !exist {
					t.Fatalf("Project %v was not found after setting the due date", tc.project.Id)
				}
				if updatedProject.(Project).Due != tc.dueDate {
					t.Errorf("Expected due date to be %v, got %v", tc.dueDate, updatedProject.(Project).Due)
				}
			}
		})
	}
}

// getters

func Test_GetAllProjects(t *testing.T) {
    
    // mocking the tasks based on previous tests
    // TODO: add isolation?
    knownProjects := []Project{
        {Id: "1", Body: "First project", RealmId: "1", ContextId: "1",},
        {Id: "10", Body: "Project 10", RealmId: "2", ContextId: "2", Due: "2024-01-01"},
		{Id: "2", Body: "Edited project content", RealmId: "1", ContextId: "2",},
        {Id: "20", Body: "Removable project", RealmId: "1", ContextId: "2",},
        {Id: "21", Body: "Project 21", RealmId: "1",},
        {Id: "22", Body: "Project 22", RealmId: "1",},
		{Id: "3", Body: "Project id 3 content", RealmId: "2", ContextId: "1",},
    }

    // Manually marshal the known projects to create the expected outcome.
    projectsObject := ProjectsObject{Projects: knownProjects}
    expected, err := projectsObject.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal known projects: %v", err)
    }

    // Execute GetAllProjects() to get the actual outcome.
    actual, err := GetAllProjects()
    if err != nil {
        t.Fatalf("GetAllProjects() failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual project JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}

func Test_GetProjectsByDate(t *testing.T) {
	
	tests := []struct {
		name      string
		projectDate  string
		filterType string
		want      string
		wantErr   bool
	}{
		{"SpecificDate", "2024-01-01", "specific", `{"projects":[{"projectId":"10","projectContextId":"2","projectRealmId":"2","projectTasks":[],"projectBody":"Project 10","projectDue":"2024-01-01"}]}`, false},
		{"BeforeDate", "2022-04-05", "before", "", false},
		{"AfterDate", "2025-04-05", "after", "", false},
		{"NoMatch", "2002-04-07", "specific", "", false},
		{"InvalidDateFormat", "April 5, 2023", "specific", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProjectsByDate(tt.projectDate, tt.filterType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectsByDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("GetProjectsByDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetProjectTasks(t *testing.T){
    
    task := Task{Id: "1", RealmId: "1", Body: "First task", ContextId: "1",}

    project, perr := GetProjectById("1")
    if perr != nil {
        t.Errorf("GetProjectById() failed, %v", perr)
    }

    // test attaching to an existing project
    atterr := task.AttachTaskToProject(project)
    if atterr != nil {
        t.Errorf("AttachTaskToProject() failed, %v", atterr)
    }

    projectTasks, pterr := project.GetProjectTasks()
    if len(projectTasks) == 0 {
        t.Errorf("GetProjectTasks() failed, %v", pterr)
    }

    // test detaching from an existing project
    dtterr := task.DetachTaskFromProject(project)
    if dtterr != nil {
        t.Errorf("DetachTaskFromProject() failed, %v", dtterr)
    }

    projectWithNoTasks, pterr := project.GetProjectTasks()
    if len(projectWithNoTasks) != 0 {
        t.Errorf("GetProjectTasks() after detach failed, %v", pterr)
    }
}

func Test_GetProjectById(t *testing.T){
    // test getting a non-existing project
    nonProject, err := GetProjectById("0")
    if err != ErrProjectByIdNotFound {
        t.Fatalf("Expected ErrProjectByIdNotFound, got: %v", err)
    }

    // test getting the correct task by id
    correctProject, err := GetProjectById("1")
    if err != nil {
        t.Fatalf("Failed to get project by id, error: %v", err)
    }

    if correctProject.Body != "First project" {
        t.Fatalf("Got the wrong project, with body: %v", correctProject.Body)
    }
}

func Test_GetProjectsByRealm(t *testing.T) {
    
    // mocking the projects based on previous tests
    // TODO: add isolation?
    projectsInAssessRealm := []Project{
        {Id: "1", Body: "First project", RealmId: "1", ContextId: "1",},
		{Id: "2", Body: "Edited project content", RealmId: "1", ContextId: "2",},
        {Id: "20", Body: "Removable project", RealmId: "1", ContextId: "2",},
        {Id: "21", Body: "Project 21", RealmId: "1",},
        {Id: "22", Body: "Project 22", RealmId: "1",},
    }

    // Manually marshal the known projects to create the expected outcome.
    projectsObjectAssess := ProjectsObject{Projects: projectsInAssessRealm}
    expected, err := projectsObjectAssess.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal projects in Assess: %v", err)
    }

    actual, err := GetProjectsByRealm("1")
    if err != nil {
        t.Fatalf("GetProjectByRealm('1') failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual projects JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}

func Test_GetProjectsByContext(t *testing.T) {
    
    // mocking the projects based on previous tests
    // TODO: add isolation?
    projectsInContextOne := []Project{
        {Id: "1", Body: "First project", RealmId: "1", ContextId: "1",},
		{Id: "3", Body: "Project id 3 content", RealmId: "2", ContextId: "1",},
    }

    // Manually marshal the known tasks to create the expected outcome.
    projectsObjectForContexts := ProjectsObject{Projects: projectsInContextOne}
    expected, err := projectsObjectForContexts.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal projects for ContextId 1: %v", err)
    }

    actual, err := GetProjectsByContext("1")
    if err != nil {
        t.Fatalf("GetProjectsByContext('1') failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual project JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}
*/

