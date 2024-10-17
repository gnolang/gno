package zentasktic

import (
	"testing"

    "gno.land/p/demo/avl"
)
/*
func Test_AddContext(t *testing.T) {
    
    context := Context{Id: "1", Name: "Work"}

    // Test adding a context successfully.
    err := context.AddContext()
    if err != nil {
        t.Errorf("Failed to add context: %v", err)
    }

    // Test adding a duplicate task.
    cerr := context.AddContext()
    if cerr != ErrContextIdAlreadyExists {
        t.Errorf("Expected ErrContextIdAlreadyExists, got %v", cerr)
    }
}

func Test_EditContext(t *testing.T) {
    
    context := Context{Id: "2", Name: "Home"}

    // Test adding a context successfully.
    err := context.AddContext()
    if err != nil {
        t.Errorf("Failed to add context: %v", err)
    }

    // Test editing the context
    editedContext := Context{Id: "2", Name: "Shopping"}
    cerr := editedContext.EditContext()
    if cerr != nil {
        t.Errorf("Failed to edit the context")
    }

    retrievedContext, _ := GetContextById(editedContext.Id)
    if retrievedContext.Name != "Shopping" {
        t.Errorf("Context was not edited")
    }
}

func Test_RemoveContext(t *testing.T) {
    
    context := Context{Id: "4", Name: "Gym",}

    // Test adding a context successfully.
    err := context.AddContext()
    if err != nil {
        t.Errorf("Failed to add context: %v", err)
    }

    retrievedContext, rerr := GetContextById(context.Id)
    if rerr != nil {
        t.Errorf("Could not retrieve the added context")
    }
    // Test removing a context
    cerr := retrievedContext.RemoveContext()
    if cerr != ErrContextNotRemoved {
        t.Errorf("Expected ErrContextNotRemoved, got %v", cerr)
    }
}

func Test_AddContextToTask(t *testing.T) {

    task := Task{Id: "10", Body: "First content", RealmId: "2", ContextId: "1",}

    // Test adding a task successfully.
    err := task.AddTask()
    if err != nil {
        t.Errorf("Failed to add task: %v", err)
    }
    taskInDecide, exist := Tasks.Get("10")
	if !exist {
		t.Errorf("Task with id 10 not found")
	}
	// check if context exists
	contextToAdd, cexist := Contexts.Get("2")
	if !cexist {
		t.Errorf("Context with id 2 not found")
	}

    derr := contextToAdd.(Context).AddContextToTask(taskInDecide.(Task))
    if derr != nil {
        t.Errorf("Could not add context to a task in Decide, err %v", derr)
    }
}

func Test_AddContextToProject(t *testing.T) {

    project := Project{Id: "10", Body: "Project 10", RealmId: "2", ContextId: "1",}

    // Test adding a project successfully.
    err := project.AddProject()
    if err != nil {
        t.Errorf("Failed to add project: %v", err)
    }
    projectInDecide, exist := Projects.Get("10")
	if !exist {
		t.Errorf("Project with id 10 not found")
	}
	// check if context exists
	contextToAdd, cexist := Contexts.Get("2")
	if !cexist {
		t.Errorf("Context with id 2 not found")
	}

    derr := contextToAdd.(Context).AddContextToProject(projectInDecide.(Project))
    if derr != nil {
        t.Errorf("Could not add context to a project in Decide, err %v", derr)
    }
}

func Test_GetAllContexts(t *testing.T) {
    
    // mocking the contexts based on previous tests
    // TODO: add isolation?
    knownContexts := []Context{
        {Id: "1", Name: "Work",},
        {Id: "2", Name: "Shopping",},
        {Id: "4", Name: "Gym",},
    }

    // Manually marshal the known contexts to create the expected outcome.
    contextsObject := ContextsObject{Contexts: knownContexts}
    expected, err := contextsObject.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal known contexts: %v", err)
    }

    // Execute GetAllContexts() to get the actual outcome.
    actual, err := GetAllContexts()
    if err != nil {
        t.Fatalf("GetAllContexts() failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual contexts JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}
*/

