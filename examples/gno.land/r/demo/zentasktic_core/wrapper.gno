package zentasktic_core


import (
	"strconv"

	"gno.land/p/demo/zentasktic"
)

// this is a convenience wrapper on top of the functions declared in the zentasktic package
// to maintain consistency

// wrapping zentasktic types

type WorkableTask struct {
    zentasktic.Task
}
type WorkableProject struct {
    zentasktic.Project
}

type WorkableRealm struct {
	Id   string
	Name string
}

type WorkableContext struct {
	zentasktic.Context
}

type WorkableCollection struct {
	zentasktic.Collection
}

type WorkableObjectPath struct {
	zentasktic.ObjectPath
}

// zentasktic managers

var ztm *zentasktic.ZTaskManager
var zpm *zentasktic.ZProjectManager
var zrm *zentasktic.ZRealmManager
var zcm *zentasktic.ZContextManager
var zcl *zentasktic.ZCollectionManager
var zom *zentasktic.ZObjectPathManager
var currentTaskID int
var currentProjectTaskID int
var currentProjectID int
var currentContextID int
var currentCollectionID int
var currentPathID int

func init() {
    ztm = zentasktic.NewZTaskManager()
    zpm = zentasktic.NewZProjectManager()
	zrm = zentasktic.NewZRealmManager()
	zcm = zentasktic.NewZContextManager()
	zcl = zentasktic.NewZCollectionManager()
	zom = zentasktic.NewZObjectPathManager()
	currentTaskID = 0
	currentProjectTaskID = 0
	currentProjectID = 0
	currentContextID = 0
	currentCollectionID = 0
	currentPathID = 0
}

// tasks

func AddTask(taskBody string) error {
	taskID := incrementTaskID()
	wt := &WorkableTask{
		Task: zentasktic.Task{
			Id:          strconv.Itoa(taskID),
			Body:        taskBody,
			RealmId: 	 "1",
		},
	}
	return ztm.AddTask(wt.Task)
}


func EditTask(taskId string, taskBody string) error {
	taskToEdit, err := GetTaskById(taskId)
	if err != nil {
		return err	
	}
	taskToEdit.Body = taskBody;
	return ztm.EditTask(taskToEdit.Task)
}

func RemoveTask(taskId string) error {
	taskToRemove, err := GetTaskById(taskId)
	if err != nil {
		return err
	}
	return ztm.RemoveTask(taskToRemove.Task)
}

func MoveTaskToRealm(taskId string, realmId string) error {
	return ztm.MoveTaskToRealm(taskId, realmId)
}

func SetTaskDueDate(taskId string, dueDate string) error {
	return ztm.SetTaskDueDate(taskId, dueDate)
}

func SetTaskAlert(taskId string, alert string) error {
	return ztm.SetTaskAlert(taskId, alert)
}

func AttachTaskToProject(taskBody string, projectId string) error {
	projectTaskID := incrementProjectTaskID()
	wt := &WorkableTask{
		Task: zentasktic.Task{
			Id:          strconv.Itoa(projectTaskID),
			Body:        taskBody,
			RealmId: 	 "1",
		},
	}
	//ztm.AddTask(wt.Task)
	projectToAdd, err := GetProjectById(projectId)
	if err != nil {
		return err	
	}
	return zpm.AttachTaskToProject(ztm, wt.Task, projectToAdd.Project)
}

func EditProjectTask(projectTaskId string, projectTaskBody string, projectId string) error {
	return zpm.EditProjectTask(projectTaskId, projectTaskBody, projectId)
}

func DetachTaskFromProject(projectTaskId string, projectId string) error {
	projectToDetachFrom, err := GetProjectById(projectId)
	if err != nil {
		return err	
	}
	detachedTaskId := strconv.Itoa(incrementTaskID())
	return zpm.DetachTaskFromProject(ztm, projectTaskId, detachedTaskId, projectToDetachFrom.Project)
}

func RemoveTaskFromProject(projectTaskId string, projectId string) error {
	return zpm.RemoveTaskFromProject(projectTaskId, projectId)
}

func GetTaskById(taskId string) (WorkableTask, error) {
	task, err := ztm.GetTaskById(taskId)
	if err != nil {
		return WorkableTask{}, err
	}
	return WorkableTask{Task: task}, nil
}

func GetAllTasks() (string){
	return ztm.GetAllTasks()
}

func GetTasksByRealm(realmId string) (string){
	return ztm.GetTasksByRealm(realmId)
}

func GetTasksByContextAndRealm(contextId string, realmId string) (string){
	return ztm.GetTasksByContextAndRealm(contextId, realmId)
}

func GetTasksByDate(dueDate string, filterType string) (string){
	return ztm.GetTasksByDate(dueDate, filterType)
}

func incrementTaskID() int {
	currentTaskID++
	return currentTaskID
}

func incrementProjectTaskID() int {
	currentProjectTaskID++
	return currentProjectTaskID
}

// projects

func AddProject(projectBody string) error {
	projectID := incrementProjectID()
	wp := &WorkableProject{
		Project: zentasktic.Project{
			Id:          strconv.Itoa(projectID),
			Body:        projectBody,
			RealmId: 	 "1",
		},
	}
	return zpm.AddProject(wp.Project)
}

func EditProject(projectId string, projectBody string) error {
	projectToEdit, err := GetProjectById(projectId)
	if err != nil {
		return err	
	}
	projectToEdit.Body = projectBody;
	return zpm.EditProject(projectToEdit.Project)
}

func RemoveProject(projectId string) error {
	projectToRemove, err := GetProjectById(projectId)
	if err != nil {
		return err
	}
	return zpm.RemoveProject(projectToRemove.Project)
}

func MoveProjectToRealm(projectId string, realmId string) error {
	return zpm.MoveProjectToRealm(projectId, realmId)
}

func MarkProjectTaskAsDone(projectId string, projectTaskId string) error {
	return zpm.MarkProjectTaskAsDone(projectId, projectTaskId)
}

func GetProjectTasks(wp WorkableProject) ([]WorkableTask, error){
	tasks, err := zpm.GetProjectTasks(wp.Project)
	if err != nil {
		return nil, err
	}

	// Convert []zentasktic.Task to []WorkableTask
	var workableTasks []WorkableTask
	for _, task := range tasks {
		workableTasks = append(workableTasks, WorkableTask{Task: task})
	}

	return workableTasks, nil
}

func SetProjectDueDate(projectId string, dueDate string) error {
	return zpm.SetProjectDueDate(projectId, dueDate)
}

func GetProjectById(projectId string) (WorkableProject, error) {
	project, err := zpm.GetProjectById(projectId)
	if err != nil {
		return WorkableProject{}, err
	}
	return WorkableProject{Project: project}, nil
}

func SetProjectTaskDueDate(projectId string, projectTaskId string, dueDate string) error {
	return zpm.SetProjectTaskDueDate(projectId, projectTaskId, dueDate)
}

func GetAllProjects() (string){
	return zpm.GetAllProjects()
}

func GetProjectsByRealm(realmId string) (string){
	return zpm.GetProjectsByRealm(realmId)
}

func GetProjectsByContextAndRealm(contextId string, realmId string) (string){
	return zpm.GetProjectsByContextAndRealm(contextId, realmId)
}

func GetProjectsByDate(dueDate string, filterType string) (string){
	return zpm.GetProjectsByDate(dueDate, filterType)
}

func incrementProjectID() int {
	currentProjectID++
	return currentProjectID
}

// realms

func AddRealm(wr WorkableRealm) error {
	r := zentasktic.Realm{
		Id:   wr.Id,
		Name: wr.Name,
	}
	return zrm.AddRealm(r)
}

func RemoveRealm(wr WorkableRealm) error {
	r := zentasktic.Realm{
		Id:   wr.Id,
		Name: wr.Name,
	}
	return zrm.RemoveRealm(r)
}

func GetRealmById(realmId string) (WorkableRealm, error) {
	r, err := zrm.GetRealmById(realmId)
	if err != nil {
		return WorkableRealm{}, err
	}
	return WorkableRealm{
		Id:   r.Id,
		Name: r.Name,
	}, nil
}

func GetAllRealms() (string, error) {
	return zrm.GetRealms()
}

// contexts

func AddContext(contextName string) error {
	contextID := incrementContextID()
	wc := &WorkableContext{
		Context: zentasktic.Context{
			Id:          strconv.Itoa(contextID),
			Name:        contextName,
		},
	}
	return zcm.AddContext(wc.Context)
}

func EditContext(contextId string, newContext string) error {
	contextToEdit, err := GetContextById(contextId)
	if err != nil {
		return err	
	}
	contextToEdit.Name = newContext;
	return zcm.EditContext(contextToEdit.Context)
}

func RemoveContext(contextId string) error {
	contextToRemove, err := GetContextById(contextId)
	if err != nil {
		return err
	}
	return zcm.RemoveContext(contextToRemove.Context)
}

func AddContextToTask(contextId string, taskId string) error {
	contextToAdd, err := GetContextById(contextId)
	if err != nil {
		return err
	}
	taskToAddContextTo, merr := GetTaskById(taskId)
	if merr != nil {
		return merr
	}
	return zcm.AddContextToTask(ztm, contextToAdd.Context, taskToAddContextTo.Task)
}

func AddContextToProject(contextId string, projectId string) error {
	contextToAdd, err := GetContextById(contextId)
	if err != nil {
		return err
	}
	projectToAddContextTo, merr := GetProjectById(projectId)
	if merr != nil {
		return merr
	}
	return zcm.AddContextToProject(zpm, contextToAdd.Context, projectToAddContextTo.Project)
}

func AddContextToProjectTask(contextId string, projectId string, projectTaskId string) error {
	contextToAdd, err := GetContextById(contextId)
	if err != nil {
		return err
	}
	projectToAddContextTo, merr := GetProjectById(projectId)
	if merr != nil {
		return merr
	}
	return zcm.AddContextToProjectTask(zpm, contextToAdd.Context, projectToAddContextTo.Project, projectTaskId)
}

func GetContextById(contextId string) (WorkableContext, error) {
	context, err := zcm.GetContextById(contextId)
	if err != nil {
		return WorkableContext{}, err
	}
	return WorkableContext{Context: context}, nil
}

func GetAllContexts() (string) {
	return zcm.GetAllContexts()
}

func incrementContextID() int {
	currentContextID++
	return currentContextID
}

// collections
/*
func AddCollection(wc WorkableCollection) error {
	c := zentasktic.Collection{
		Id:       wc.Id,
		RealmId:  wc.RealmId,
		Name:     wc.Name,
		Tasks:    toZentaskticTasks(wc.Tasks),
		Projects: toZentaskticProjects(wc.Projects),
	}
	return zcl.AddCollection(c)
}

func EditCollection(wc WorkableCollection) error {
	c := zentasktic.Collection{
		Id:       wc.Id,
		RealmId:  wc.RealmId,
		Name:     wc.Name,
		Tasks:    toZentaskticTasks(wc.Tasks),
		Projects: toZentaskticProjects(wc.Projects),
	}
	return zcl.EditCollection(c)
}

func RemoveCollection(wc WorkableCollection) error {
	c := zentasktic.Collection{
		Id:       wc.Id,
		RealmId:  wc.RealmId,
		Name:     wc.Name,
		Tasks:    toZentaskticTasks(wc.Tasks),
		Projects: toZentaskticProjects(wc.Projects),
	}
	return zcl.RemoveCollection(c)
}

func GetCollectionById(collectionId string) (WorkableCollection, error) {
	c, err := zcl.GetCollectionById(collectionId)
	if err != nil {
		return WorkableCollection{}, err
	}
	return WorkableCollection{
		Id:       c.Id,
		RealmId:  c.RealmId,
		Name:     c.Name,
		Tasks:    toWorkableTasks(c.Tasks),
		Projects: toWorkableProjects(c.Projects),
	}, nil
}

func GetCollectionTasks(wc WorkableCollection) ([]WorkableTask, error) {
	c := zentasktic.Collection{
		Id: wc.Id,
	}
	tasks, err := zcl.GetCollectionTasks(c)
	if err != nil {
		return nil, err
	}
	return toWorkableTasks(tasks), nil
}

func GetCollectionProjects(wc WorkableCollection) ([]WorkableProject, error) {
	c := zentasktic.Collection{
		Id: wc.Id,
	}
	projects, err := zcl.GetCollectionProjects(c)
	if err != nil {
		return nil, err
	}
	return toWorkableProjects(projects), nil
}

func GetAllCollections() (string, error) {
	return zcl.GetAllCollections()
}

// Helper functions to convert between Workable and zentasktic types
func toZentaskticTasks(tasks []WorkableTask) []zentasktic.Task {
	ztasks := make([]zentasktic.Task, len(tasks))
	for i, t := range tasks {
		ztasks[i] = t.Task
	}
	return ztasks
}

func toWorkableTasks(tasks []zentasktic.Task) []WorkableTask {
	wtasks := make([]WorkableTask, len(tasks))
	for i, t := range tasks {
		wtasks[i] = WorkableTask{Task: t}
	}
	return wtasks
}

func toZentaskticProjects(projects []WorkableProject) []zentasktic.Project {
	zprojects := make([]zentasktic.Project, len(projects))
	for i, p := range projects {
		zprojects[i] = p.Project
	}
	return zprojects
}

func toWorkableProjects(projects []zentasktic.Project) []WorkableProject {
	wprojects := make([]WorkableProject, len(projects))
	for i, p := range projects {
		wprojects[i] = WorkableProject{Project: p}
	}
	return wprojects
}*/

// object Paths

func AddPath(wop WorkableObjectPath) error {
	o := zentasktic.ObjectPath{
		ObjectType: wop.ObjectType,
		Id:         wop.Id,
		RealmId:    wop.RealmId,
	}
	return zom.AddPath(o)
}


func GetObjectJourney(objectType string, objectId string) (string, error) {
	return zom.GetObjectJourney(objectType, objectId)
}