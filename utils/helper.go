package utils

import (
	"encoding/json"
	"log"
)

type STIXObject struct {
	Type               string        `json:"type"`
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Aliases            []string      `json:"aliases"`
	ExternalReferences []ExternalRef `json:"external_references"`
	XMitreDeprecated   bool          `json:"x_mitre_deprecated"`
	Revoked            bool          `json:"revoked"`
}

type ExternalRef struct {
	ExternalID string `json:"external_id"`
}

func IsDeprecatedOrRevoked(obj map[string]interface{}) bool {
	if deprecated, ok := obj["x_mitre_deprecated"].(bool); ok && deprecated {
		return true
	}
	if revoked, ok := obj["revoked"].(bool); ok && revoked {
		return true
	}
	return false
}

func ConvertToStringSlice(interfaces []interface{}) []string {
	var strings []string
	for _, iface := range interfaces {
		strings = append(strings, iface.(string))
	}
	return strings
}

func ToSTIXObject(obj map[string]interface{}) STIXObject {
	var stixObj STIXObject
	objJSON, _ := json.Marshal(obj)
	err := json.Unmarshal(objJSON, &stixObj)
	if err != nil {
		log.Fatalf("Unable to marchel STIX Object json: %v\n", err)
	}
	return stixObj
}
