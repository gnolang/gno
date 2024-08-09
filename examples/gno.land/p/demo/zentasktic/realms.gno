// base implementation
package zentasktic

import (
	"gno.land/p/demo/avl"
)

// structs

type Realm struct {
	Id 			string `json:"realmId"`
	Name 		string `json:"realmName"`
}

type ZRealmManager struct {
	Realms *avl.Tree
}

func NewZRealmManager() *ZRealmManager {
	zrm := &ZRealmManager{
		Realms: avl.NewTree(),
	}
	zrm.initializeHardcodedRealms()
	return zrm
}


func (zrm *ZRealmManager) initializeHardcodedRealms() {
	hardcodedRealms := []Realm{
		{Id: "1", Name: "Assess"},
		{Id: "2", Name: "Decide"},
		{Id: "3", Name: "Do"},
		{Id: "4", Name: "Collections"},
	}

	for _, realm := range hardcodedRealms {
		zrm.Realms.Set(realm.Id, realm)
	}
}


func (zrm *ZRealmManager) AddRealm(r Realm) (err error){
	// implementation
	if zrm.Realms.Size() != 0 {
		_, exist := zrm.Realms.Get(r.Id)
		if exist {
			return ErrRealmIdAlreadyExists
		}
	}
	// check for hardcoded values
	if r.Id == "1" || r.Id == "2" || r.Id == "3" || r.Id == "4" {
		return ErrRealmIdNotAllowed
	}
	zrm.Realms.Set(r.Id, r)
	return nil
	
}

func (zrm *ZRealmManager) RemoveRealm(r Realm) (err error){
	// implementation
	if zrm.Realms.Size() != 0 {
		_, exist := zrm.Realms.Get(r.Id)
		if !exist {
			return ErrRealmIdNotFound
		} else {
			// check for hardcoded values, not removable
			if r.Id == "1" || r.Id == "2" || r.Id == "3" || r.Id == "4" {
				return ErrRealmIdNotAllowed
			}
		}
	}
	
	_, removed := zrm.Realms.Remove(r.Id)
	if !removed {
		return ErrRealmNotRemoved
	}
	return nil
	
}

// getters
func (zrm *ZRealmManager) GetRealmById(realmId string) (r Realm, err error) {
	// implementation
	if zrm.Realms.Size() != 0 {
		rInterface, exist := zrm.Realms.Get(realmId)
		if exist {
			return rInterface.(Realm), nil
		} else {
			return Realm{}, ErrRealmIdNotFound
		}
	}
	return Realm{}, ErrRealmIdNotFound
}

func (zrm *ZRealmManager) GetRealms() (realms string, err error) {
	// implementation
	var allRealms []Realm

	// Iterate over the Realms AVL tree to collect all Context objects.
	zrm.Realms.Iterate("", "", func(key string, value interface{}) bool {
		if realm, ok := value.(Realm); ok {
			allRealms = append(allRealms, realm)
		}
		return false // Continue iteration until all nodes have been visited.
	})


	// Create a RealmsObject with all collected contexts.
	realmsObject := &RealmsObject{
		Realms: allRealms,
	}

	// Use the custom MarshalJSON method to marshal the realms into JSON.
	marshalledRealms, rerr := realmsObject.MarshalJSON()
	if rerr != nil {
		return "", rerr
	} 
	return string(marshalledRealms), nil
}
