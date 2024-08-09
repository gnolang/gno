// base implementation
package zentasktic

import (
	"strconv"

	"gno.land/p/demo/avl"
)

// holding the path of an object since creation
// each time we move an object from one realm to another, we add to its path
type ObjectPath struct {
	ObjectType string `json:"objectType"` // Task, Project
	Id         string `json:"id"`         // this is the Id of the object moved, Task, Project
	RealmId    string `json:"realmId"`
}

type ZObjectPathManager struct {
	Paths avl.Tree
	PathId int
}

func NewZObjectPathManager() *ZObjectPathManager {
	return &ZObjectPathManager{
		Paths: *avl.NewTree(),
		PathId: 1,
	}
}

func (zopm *ZObjectPathManager) AddPath(o ObjectPath) error {
	zopm.PathId++
	updated := zopm.Paths.Set(strconv.Itoa(zopm.PathId), o)
	if !updated {
		return ErrObjectPathNotUpdated
	}
	return nil
}

func (zopm *ZObjectPathManager) GetObjectJourney(objectType string, objectId string) (string, error) {
	var objectPaths []ObjectPath

	// Iterate over the Paths AVL tree to collect all ObjectPath objects.
	zopm.Paths.Iterate("", "", func(key string, value interface{}) bool {
		if objectPath, ok := value.(ObjectPath); ok {
			if objectPath.ObjectType == objectType && objectPath.Id == objectId {
				objectPaths = append(objectPaths, objectPath)
			}
		}
		return false // Continue iteration until all nodes have been visited.
	})

	// Create an ObjectJourney with all collected paths.
	objectJourney := &ObjectJourney{
		ObjectPaths: objectPaths,
	}

	// Use the custom MarshalJSON method to marshal the journey into JSON.
	marshalledJourney, merr := objectJourney.MarshalJSON()
	if merr != nil {
		return "", merr
	}
	return string(marshalledJourney), nil
}


// GetZenStatus
/* todo: leave it to the client
func () GetZenStatus() (zenStatus string, err error) {
	// implementation
}
*/
