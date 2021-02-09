package test

import (
	"auto-portal-auth/component/utils"
	"reflect"
	"testing"
)

func TestRemoveDuplicateStrings(t *testing.T) {

	duplicateStringList := []string{
		"b4:2e:99:c4:d2:b4",
		"b4:2e:99:c4:d2:b4",
		"aa:bb:cc:dd:ee:ff",
		"aa:bb:cc:dd:ee:ff",
		"11:22:33:44:55:66",
	}

	resultStringList := []string{
		"b4:2e:99:c4:d2:b4",
		"aa:bb:cc:dd:ee:ff",
		"11:22:33:44:55:66",
	}

	resultStringMap := map[string]struct{}{
		"11:22:33:44:55:66": {},
		"aa:bb:cc:dd:ee:ff": {},
		"b4:2e:99:c4:d2:b4": {},
	}

	resList, resMap := utils.RemoveDuplicateStrings(duplicateStringList)
	if !reflect.DeepEqual(resList, resultStringList) {
		t.Error("Error removing duplicated elements and refactor a new list")
	}
	if !reflect.DeepEqual(resMap, resultStringMap) {
		t.Error("Error removing duplicated elements and refactor a new map")
	}

}
