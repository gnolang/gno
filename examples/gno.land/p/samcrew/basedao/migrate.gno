package basedao

import (
	"errors"

	"gno.land/p/samcrew/daokit"
)

// Making this action name extra explicit.
const ActionChangeDAOImplementationKind = "gno.land/p/samcrew/basedao.ChangeDAOImplementation"

type MigrateFn = func(prev *DAOPrivate, params []any) daokit.DAO

type MigrationParamsFn = func() []any

type actionChangeDAOImplementation struct {
	migrate MigrateFn
}

func (a *actionChangeDAOImplementation) String() string {
	return "WARNING: thoroughly check the migration code before approving this"
}

func NewChangeDAOImplementationHandler(dao *DAOPrivate, setImplem daokit.SetImplemFn, paramsFn MigrationParamsFn) daokit.ActionHandler {
	if dao == nil || setImplem == nil || paramsFn == nil {
		panic("nil arg")
	}
	return daokit.NewActionHandler(ActionChangeDAOImplementationKind, func(i interface{}) {
		action, ok := i.(*actionChangeDAOImplementation)
		if !ok {
			panic(errors.New("invalid action type"))
		}

		migrated := action.migrate(dao, paramsFn())
		setImplem(migrated)
	})
}

func NewChangeDAOImplementationAction(migrate MigrateFn) daokit.Action {
	return daokit.NewAction(ActionChangeDAOImplementationKind, &actionChangeDAOImplementation{migrate: migrate})
}
