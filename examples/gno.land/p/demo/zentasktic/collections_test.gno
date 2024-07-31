package zentasktic

import (
	"testing"

    "gno.land/p/demo/avl"
)
/*

func Test_AddCollection(t *testing.T) {
    
    collection := Collection{Id: "1", RealmId: "4", Name: "First collection",}

    // Test adding a collection successfully.
    err := collection.AddCollection()
    if err != nil {
        t.Errorf("Failed to add collection: %v", err)
    }

    // Test adding a duplicate task.
    cerr := collection.AddCollection()
    if cerr != ErrCollectionIdAlreadyExists {
        t.Errorf("Expected ErrCollectionIdAlreadyExists, got %v", cerr)
    }
}

func Test_RemoveCollection(t *testing.T) {
    
    collection := Collection{Id: "20", RealmId: "4", Name: "Removable collection",}

    // Test adding a collection successfully.
    err := collection.AddCollection()
    if err != nil {
        t.Errorf("Failed to add collection: %v", err)
    }

    retrievedCollection, rerr := GetCollectionById(collection.Id)
    if rerr != nil {
        t.Errorf("Could not retrieve the added collection")
    }

    // Test removing a collection
    terr := retrievedCollection.RemoveCollection()
    if terr != ErrCollectionNotRemoved {
        t.Errorf("Expected ErrCollectionNotRemoved, got %v", terr)
    }
}

func Test_EditCollection(t *testing.T) {
    
    collection := Collection{Id: "2", RealmId: "4", Name: "Second collection",}

    // Test adding a collection successfully.
    err := collection.AddCollection()
    if err != nil {
        t.Errorf("Failed to add collection: %v", err)
    }

    // Test editing the collection
    editedCollection := Collection{Id: collection.Id, RealmId: collection.RealmId, Name: "Edited collection",}
    cerr := editedCollection.EditCollection()
    if cerr != nil {
        t.Errorf("Failed to edit the collection")
    }

    retrievedCollection, _ := GetCollectionById(editedCollection.Id)
    if retrievedCollection.Name != "Edited collection" {
        t.Errorf("Collection was not edited")
    }
}

func Test_AddProjectToCollection(t *testing.T){
    // Example Collection and Projects
    col := Collection{Id: "1", Name: "First collection", RealmId: "4",}
    prj := Project{Id: "10", Body: "Project 10", RealmId: "1",}

    Collections.Set(col.Id, col) // Mock existing collections

    tests := []struct {
        name    string
        collection Collection
        project    Project
        wantErr bool
        errMsg  error
    }{
        {
            name:    "Attach to existing collection",
            collection: col,
            project:    prj,
            wantErr: false,
        },
        {
            name:    "Attach to non-existing collection",
            collection: Collection{Id: "200", Name: "Collection 200", RealmId: "4",},
            project:    prj,
            wantErr: true,
            errMsg:  ErrCollectionIdNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.collection.AddProjectToCollection(tt.project)
            if (err != nil) != tt.wantErr {
                t.Errorf("AddProjectToCollection() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && err != tt.errMsg {
                t.Errorf("AddProjectToCollection() error = %v, expected %v", err, tt.errMsg)
            }

            // For successful attach, verify the project is added to the collection's tasks.
            if !tt.wantErr {
                projects, exist := CollectionProjects.Get(tt.collection.Id)
                if !exist || len(projects.([]Project)) == 0 {
                    t.Errorf("Project was not added to the collection")
                } else {
                    found := false
                    for _, project := range projects.([]Project) {
                        if project.Id == tt.project.Id {
                            found = true
                            break
                        }
                    }
                    if !found {
                        t.Errorf("Project was not attached to the collection")
                    }
                }
            }
        })
    }
}

func Test_AddTaskToCollection(t *testing.T){
    // Example Collection and Tasks
    col := Collection{Id: "2", Name: "Second Collection", RealmId: "4",}
    tsk := Task{Id: "30", Body: "Task 30", RealmId: "1",}

    Collections.Set(col.Id, col) // Mock existing collections

    tests := []struct {
        name    string
        collection Collection
        task    Task
        wantErr bool
        errMsg  error
    }{
        {
            name:    "Attach to existing collection",
            collection: col,
            task:    tsk,
            wantErr: false,
        },
        {
            name:    "Attach to non-existing collection",
            collection: Collection{Id: "210", Name: "Collection 210", RealmId: "4",},
            task:    tsk,
            wantErr: true,
            errMsg:  ErrCollectionIdNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.collection.AddTaskToCollection(tt.task)
            if (err != nil) != tt.wantErr {
                t.Errorf("AddTaskToCollection() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && err != tt.errMsg {
                t.Errorf("AddTaskToCollection() error = %v, expected %v", err, tt.errMsg)
            }

            // For successful attach, verify the task is added to the collection's tasks.
            if !tt.wantErr {
                tasks, exist := CollectionTasks.Get(tt.collection.Id)
                if !exist || len(tasks.([]Task)) == 0 {
                    t.Errorf("Task was not added to the collection")
                } else {
                    found := false
                    for _, task := range tasks.([]Task) {
                        if task.Id == tt.task.Id {
                            found = true
                            break
                        }
                    }
                    if !found {
                        t.Errorf("Task was not attached to the collection")
                    }
                }
            }
        })
    }
}

func Test_RemoveProjectFromCollection(t *testing.T){
    // Setup:
	collection := Collection{Id: "300", Name: "Collection 300",}
	project1 := Project{Id: "21", Body: "Project 21", RealmId: "1",}
	project2 := Project{Id: "22", Body: "Project 22", RealmId: "1",}

    collection.AddCollection()
    project1.AddProject()
    project2.AddProject()
    collection.AddProjectToCollection(project1)
    collection.AddProjectToCollection(project2)

	tests := []struct {
		name       string
		project    Project
		collection    Collection
		wantErr    bool
		expectedErr error
	}{
		{
			name:       "Remove existing project from collection",
			project:       project1,
			collection:    collection,
			wantErr:    false,
			expectedErr: nil,
		},
		{
			name:       "Try to remove project from non-existing collection",
			project:       project1,
			collection:    Collection{Id: "nonexistent"},
			wantErr:    true,
			expectedErr: ErrCollectionIdNotFound,
		},
		{
			name:       "Try to remove non-existing project from collection",
			project:       Project{Id: "nonexistent"},
			collection:    collection,
			wantErr:    true,
			expectedErr: ErrProjectByIdNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.RemoveProjectFromCollection(tt.project)

			if tt.wantErr {
				if err == nil || err != tt.expectedErr {
					t.Errorf("%s: expected error %v, got %v", tt.name, tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}

				// For successful removal, verify the project is no longer part of the collection's projects
				if !tt.wantErr {
					projects, _ := CollectionProjects.Get(tt.collection.Id)
					for _, project := range projects.([]Project) {
						if project.Id == tt.project.Id {
							t.Errorf("%s: project was not detached from the collection", tt.name)
							break
						}
					}
				}
			}
		})
	}
}

func Test_RemoveTaskFromCollection(t *testing.T){
    // setup, re-using parts from Test_AddTaskToCollection
	collection := Collection{Id: "40", Name: "Collection 40",}
    task1 := Task{Id: "40", Body: "Task 40", RealmId: "1",}

    collection.AddCollection()
    task1.AddTask()
    collection.AddTaskToCollection(task1)

	tests := []struct {
		name       string
		task    Task
		collection    Collection
		wantErr    bool
		expectedErr error
	}{
		{
			name:       "Remove existing task from collection",
			task:       task1,
			collection:    collection,
			wantErr:    false,
			expectedErr: nil,
		},
		{
			name:       "Try to remove task from non-existing collection",
			task:       task1,
			collection:    Collection{Id: "nonexistent"},
			wantErr:    true,
			expectedErr: ErrCollectionIdNotFound,
		},
		{
			name:       "Try to remove non-existing task from collection",
			task:       Task{Id: "nonexistent"},
			collection:    collection,
			wantErr:    true,
			expectedErr: ErrTaskByIdNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.RemoveTaskFromCollection(tt.task)

			if tt.wantErr {
				if err == nil || err != tt.expectedErr {
					t.Errorf("%s: expected error %v, got %v", tt.name, tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.name, err)
				}

				// For successful removal, verify the task is no longer part of the collection's tasks
				if !tt.wantErr {
					tasks, _ := CollectionTasks.Get(tt.collection.Id)
					for _, task := range tasks.([]Task) {
						if task.Id == tt.task.Id {
							t.Errorf("%s: task was not detached from the collection", tt.name)
							break
						}
					}
				}
			}
		})
	}
}

func Test_GetCollectionById(t *testing.T){
    // test getting a non-existing collection
    nonCollection, err := GetCollectionById("0")
    if err != ErrCollectionByIdNotFound {
        t.Fatalf("Expected ErrCollectionByIdNotFound, got: %v", err)
    }

    // test getting the correct collection by id
    correctCollection, err := GetCollectionById("1")
    if err != nil {
        t.Fatalf("Failed to get collection by id, error: %v", err)
    }

    if correctCollection.Name != "First collection" {
        t.Fatalf("Got the wrong collection, with name: %v", correctCollection.Name)
    }
}

func Test_GetCollectionTasks(t *testing.T) {
    // retrieving objects based on these mocks
    //col := Collection{Id: "2", Name: "Second Collection", RealmId: "4",}
    tsk := Task{Id: "30", Body: "Task 30", RealmId: "1",}

    collection, cerr := GetCollectionById("2")
    if cerr != nil {
        t.Errorf("GetCollectionById() failed, %v", cerr)
    }

    collectionTasks, pterr := collection.GetCollectionTasks()
    if len(collectionTasks) == 0 {
        t.Errorf("GetCollectionTasks() failed, %v", pterr)
    }

    // test detaching from an existing collection
    dtterr := collection.RemoveTaskFromCollection(tsk)
    if dtterr != nil {
        t.Errorf("RemoveTaskFromCollection() failed, %v", dtterr)
    }

    collectionWithNoTasks, pterr := collection.GetCollectionTasks()
    if len(collectionWithNoTasks) != 0 {
        t.Errorf("GetCollectionTasks() after detach failed, %v", pterr)
    }

    // add task back to collection, for tests mockup integrity
    collection.AddTaskToCollection(tsk)
}

func Test_GetCollectionProjects(t *testing.T) {
    // retrieving objects based on these mocks
    //col := Collection{Id: "1", Name: "First Collection", RealmId: "4",}
    prj := Project{Id: "10", Body: "Project 10", RealmId: "2", ContextId: "2", Due: "2024-01-01"}

    collection, cerr := GetCollectionById("1")
    if cerr != nil {
        t.Errorf("GetCollectionById() failed, %v", cerr)
    }

    collectionProjects, pterr := collection.GetCollectionProjects()
    if len(collectionProjects) == 0 {
        t.Errorf("GetCollectionProjects() failed, %v", pterr)
    }

    // test detaching from an existing collection
    dtterr := collection.RemoveProjectFromCollection(prj)
    if dtterr != nil {
        t.Errorf("RemoveProjectFromCollection() failed, %v", dtterr)
    }

    collectionWithNoProjects, pterr := collection.GetCollectionProjects()
    if len(collectionWithNoProjects) != 0 {
        t.Errorf("GetCollectionProjects() after detach failed, %v", pterr)
    }

    // add project back to collection, for tests mockup integrity
    collection.AddProjectToCollection(prj)
}

func Test_GetAllCollections(t *testing.T){
    // mocking the collections based on previous tests
    // TODO: add isolation?
    knownCollections := []Collection{
        {
            Id:       "1",
            RealmId:  "4",
            Name:     "First collection",
            Tasks:    nil, 
            Projects: []Project{
                {
                    Id:       "10",
                    ContextId: "2",
                    RealmId:   "4",
                    Tasks:     nil, 
                    Body:      "Project 10",
                    Due:       "2024-01-01",
                },
            },
        },
        {
            Id:       "2",
            RealmId:  "4",
            Name:     "Second Collection",
            Tasks:    []Task{
                {
                    Id:"30",
                    ProjectId:"",
                    ContextId:"",
                    RealmId:"4",
                    Body:"Task 30",
                    Due:"",
                    Alert:"",
                },
            },
            Projects: nil, 
        },
        {
            Id:"20",
            RealmId:"4",
            Name:"Removable collection",
            Tasks: nil,
            Projects: nil,
        },
        {
            Id:       "300",
            Name:     "Collection 300",
            Tasks:    nil, 
            Projects: []Project {
                {
                    Id:"22",
                    ContextId:"",
                    RealmId:"4",
                    Tasks: nil,
                    Body:"Project 22",
                    Due:"",
                },
            }, 
        },
        {
            Id:       "40",
            Name:     "Collection 40",
            Tasks:    nil, 
            Projects: nil, 
        },
    }
    

    // Manually marshal the known collections to create the expected outcome.
    collectionsObject := CollectionsObject{Collections: knownCollections}
    expected, err := collectionsObject.MarshalJSON()
    if err != nil {
        t.Fatalf("Failed to manually marshal known collections: %v", err)
    }

    // Execute GetAllCollections() to get the actual outcome.
    actual, err := GetAllCollections()
    if err != nil {
        t.Fatalf("GetAllCollections() failed with error: %v", err)
    }

    // Compare the expected and actual outcomes.
    if string(expected) != actual {
        t.Errorf("Expected and actual collections JSON strings do not match.\nExpected: %s\nActual: %s", string(expected), actual)
    }
}
*/




