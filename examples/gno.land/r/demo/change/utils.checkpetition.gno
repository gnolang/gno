package change

import (
	"errors"
	"strconv"
)

func checkPetitionId(id int) error {
	if id < 0 || id > len(petitions)-1 {
		return errors.New("Invalid petition id range (too big or too small).")
	}
	return nil
}

func checkStringPetitionId(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, checkPetitionId(id)
}

func isStringPositiveInt(intStr string) bool {
	id, err := strconv.Atoi(intStr)
	if err != nil || id <= 0 {
		return false
	}
	return true
}
