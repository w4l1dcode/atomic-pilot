package atomic

type Relationship struct {
	SourceRef string `json:"source_ref"`
	TargetRef string `json:"target_ref"`
}

func FilterRelationshipsBySource(objects []interface{}, sourceID string) []interface{} {
	var filtered []interface{}
	for _, obj := range objects {
		objMap := obj.(map[string]interface{})
		if objMap["type"] == "relationship" && objMap["source_ref"] == sourceID {
			filtered = append(filtered, obj)
		}
	}
	return filtered
}
