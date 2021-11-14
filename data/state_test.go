package data_test

import (
	"os"
	"testing"

	"github.com/danesparza/cloudjournal/data"
)

func TestConfig_UpdateLogState_ValidState_Successful(t *testing.T) {

	//	Arrange
	systemdb := getTestFiles()

	db, err := data.NewManager(systemdb)
	if err != nil {
		t.Errorf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
	}()

	testUnit := "unittest"
	testCursor := "s=7fe895b45f18448daa12dfe9ec1d2993;i=230;b=6b9d0f62f43c4b0bb0f61848b4da3b15;m=3b5c2b9;t=5cff2c0f5338d;x=274b5b63cb69c9f7"
	testNextToken := "sometokenvalue"

	//	Act
	retval, err := db.UpdateLogState(testUnit, testCursor, testNextToken)

	//	Assert
	if err != nil {
		t.Errorf("UpdateLogState - Should execute without error, but got: %s", err)
	}

	if retval.LastCursor != testCursor {
		t.Errorf("UpdateLogState - Response should match set cursor, but got: %v", retval.LastCursor)
	}

	if retval.NextSequenceToken != testNextToken {
		t.Errorf("UpdateLogState - Response should match set token, but got: %v", retval.NextSequenceToken)
	}

	if retval.LastSynced.IsZero() {
		t.Errorf("UpdateLogState failed: Should have set an item with the correct datetime: %+v", retval)
	}

}
