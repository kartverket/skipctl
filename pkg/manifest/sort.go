package manifest

import (
	"encoding/json"
	"sort"
	"strings"
)

// SortJSON parses a JSON string, sorts the resources by Kind, Namespace, and Name,
// and returns the sorted JSON string. It supports both JSON arrays and single objects (including Lists).
func SortJSON(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", nil
	}

	var resources []map[string]interface{}

	// Determine if input is array or single object
	if strings.HasPrefix(content, "[") {
		if err := json.Unmarshal([]byte(content), &resources); err != nil {
			return "", err
		}
	} else {
		var resource map[string]interface{}
		if err := json.Unmarshal([]byte(content), &resource); err != nil {
			return "", err
		}

		// Handle Kubernetes List
		if getKind(resource) == "List" {
			if items, ok := resource["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						resources = append(resources, itemMap)
					}
				}
			}
		} else {
			resources = append(resources, resource)
		}
	}

	sortResources(resources)

	// Marshal back to JSON with indentation for readable diffs
	out, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// sortResources sorts a slice of generic maps by Kind, Namespace, and Name
func sortResources(resources []map[string]interface{}) {
	sort.Slice(resources, func(i, j int) bool {
		metaI := getMetadata(resources[i])
		metaJ := getMetadata(resources[j])

		kindI := getKind(resources[i])
		kindJ := getKind(resources[j])

		if kindI != kindJ {
			return kindI < kindJ
		}
		if metaI.Namespace != metaJ.Namespace {
			return metaI.Namespace < metaJ.Namespace
		}
		return metaI.Name < metaJ.Name
	})
}

type metadata struct {
	Name      string
	Namespace string
}

func getKind(res map[string]interface{}) string {
	if v, ok := res["kind"].(string); ok {
		return v
	}
	return ""
}

func getMetadata(res map[string]interface{}) metadata {
	meta := metadata{}
	if m, ok := res["metadata"].(map[string]interface{}); ok {
		if n, ok := m["name"].(string); ok {
			meta.Name = n
		}
		if ns, ok := m["namespace"].(string); ok {
			meta.Namespace = ns
		}
	}
	return meta
}
