package volcengine

import (
	"context"
	"slices"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

const matrixKeyRegion = "region"

// BuildRegionList returns a list of matrix items, one per region specified in the connection config
func BuildRegionList(_ context.Context, d *plugin.QueryData) []map[string]interface{} {
	config := GetConfig(d.Connection)

	if config.Regions != nil {
		regions := config.Regions

		if len(getInvalidRegions(regions)) > 0 {
			panic("\n\nConnection config has invalid regions: " + strings.Join(getInvalidRegions(regions), ",") + ". Edit your connection configuration file and then restart Steampipe.")
		}

		matrix := make([]map[string]interface{}, len(regions))
		for i, region := range regions {
			matrix[i] = map[string]interface{}{matrixKeyRegion: region}
		}
		return matrix
	}

	return []map[string]interface{}{
		{matrixKeyRegion: GetDefaultRegion(d.Connection)},
	}
}

// getInvalidRegions returns a list of regions that are not valid Volcengine regions
func getInvalidRegions(regions []string) []string {
	volcengineRegions := []string{
		"cn-beijing",
		"cn-shanghai",
		"cn-guangzhou",
		"ap-southeast-1",
		"cn-beijing-autodriving",
		"cn-shanghai-autodriving",
	}

	invalidRegions := []string{}
	for _, region := range regions {
		if !slices.Contains(volcengineRegions, region) {
			invalidRegions = append(invalidRegions, region)
		}
	}
	return invalidRegions
}
