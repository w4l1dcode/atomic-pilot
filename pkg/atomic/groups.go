package atomic

import (
	"fmt"
	"math/rand"
)

func GetNextGroup(groups []interface{}) (map[string]interface{}, error) {
	var availableGroups []map[string]interface{}

	for _, group := range groups {
		groupMap, ok := group.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert group to map[string]interface{}")
		}
		availableGroups = append(availableGroups, groupMap)
	}

	selectedGroup := availableGroups[rand.Intn(len(availableGroups))]
	return selectedGroup, nil
}
