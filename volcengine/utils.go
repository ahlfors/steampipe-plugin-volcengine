package volcengine

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Constants for Standard Column Descriptions
const (
	ColumnDescriptionAkas    = "Array of globally unique identifier strings (also known as) for the resource."
	ColumnDescriptionTags    = "A map of tags for the resource."
	ColumnDescriptionTitle   = "Title of the resource."
	ColumnDescriptionAccount = "The Volcengine Account ID in which the resource is located."
	ColumnDescriptionRegion  = "The Volcengine region in which the resource is located."
)

// ensureStringArray ensures a value is returned as a string array
func ensureStringArray(_ context.Context, d *transform.TransformData) (interface{}, error) {
	switch v := d.Value.(type) {
	case []string:
		return v, nil
	case string:
		return []string{v}, nil
	default:
		str := fmt.Sprintf("%v", d.Value)
		return []string{string(str)}, nil
	}
}

// GetStringQualValue gets an equal qualifier value as string
func GetStringQualValue(quals plugin.KeyColumnQualMap, columnName string) (value *string, exists bool) {
	if quals[columnName] == nil {
		return nil, false
	}
	if quals[columnName].Quals == nil {
		return nil, false
	}

	for _, qual := range quals[columnName].Quals {
		if qual.Operator != "=" {
			return nil, true
		}
		if qual.Value != nil {
			v := qual.Value
			if v.GetListValue() != nil {
				return nil, true
			}
			s := v.GetStringValue()
			return &s, true
		}
	}
	return nil, true
}

// volcengineTagsToMap converts a list of volcengine tags (Key/Value structs) to a map
// This is a generic transform - specific tag types need specific handlers per service
func volcengineTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	// The exact type depends on the SDK struct, handle common patterns
	switch tags := d.Value.(type) {
	case []map[string]string:
		result := map[string]string{}
		for _, t := range tags {
			if k, ok := t["Key"]; ok {
				result[k] = t["Value"]
			}
		}
		return result, nil
	default:
		return d.Value, nil
	}
}
