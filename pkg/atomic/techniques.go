package atomic

import (
	"atomic-pilot/utils"
	"fmt"
	"log"
	"sync"
	"zombiezen.com/go/sqlite"
)

// dbMutex to handle concurrent access to the database.
var dbMutex sync.Mutex

func GetAllTechniques(objects []interface{}) ([]string, error) {
	techniques := make(map[string]struct{})

	for _, obj := range objects {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to assert object as map[string]interface{}")
		}
		if objMap["type"].(string) == "attack-pattern" {
			if !utils.IsDeprecatedOrRevoked(objMap) {
				techniqueObj := utils.ToSTIXObject(objMap)
				techniqueID := techniqueObj.ExternalReferences[0].ExternalID
				techniques[techniqueID] = struct{}{}
			}
		}
	}

	var allTechniques []string
	for technique := range techniques {
		allTechniques = append(allTechniques, technique)
	}
	return allTechniques, nil
}

func AreAllTechniquesUsed(conn *sqlite.Conn, allTechniques []string) bool {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	stmt := conn.Prep("SELECT COUNT(*) FROM used_techniques;")
	defer func(stmt *sqlite.Stmt) {
		if err := stmt.Finalize(); err != nil {
			log.Fatalf("Failed to finalize statement: %v\n", err)
		}
	}(stmt)

	hasRow, err := stmt.Step()
	if err != nil {
		log.Fatalf("Failed to query techniques count: %v\n", err)
	}

	if hasRow {
		usedCount := stmt.ColumnInt(0)
		return usedCount >= len(allTechniques)
	}

	return false
}

func HasTechniqueBeenUsed(conn *sqlite.Conn, technique string) bool {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	stmt := conn.Prep("SELECT COUNT(*) FROM used_techniques WHERE technique = $technique;")
	stmt.SetText("$technique", technique)
	hasRow, err := stmt.Step()
	if err != nil {
		log.Fatalf("Failed to query technique: %v\n", err)
	}
	defer func(stmt *sqlite.Stmt) {
		err := stmt.Finalize()
		if err != nil {
			log.Fatalf("Failed to close database: %v\n", err)
		}
	}(stmt)

	if hasRow && stmt.ColumnInt(0) > 0 {
		return true
	}
	return false
}

func MarkTechniqueAsUsed(conn *sqlite.Conn, technique string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	stmt := conn.Prep("INSERT INTO used_techniques (technique) VALUES ($technique);")
	stmt.SetText("$technique", technique)
	if _, err := stmt.Step(); err != nil {
		log.Fatalf("Failed to mark technique as used: %v\n", err)
	}
	defer func(stmt *sqlite.Stmt) {
		err := stmt.Finalize()
		if err != nil {
			log.Fatalf("Failed to close database: %v\n", err)
		}
	}(stmt)
}
