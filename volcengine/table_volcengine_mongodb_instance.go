package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/mongodb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineMongoDBInstance(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_mongodb_instance",
		Description: "Volcengine MongoDB Document Database Instance",
		List: &plugin.ListConfig{
			Hydrate: listMongoDBInstances,
			Tags:    map[string]string{"service": "mongodb", "action": "DescribeDBInstances"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("instance_id"),
			Hydrate:    getMongoDBInstance,
			Tags:       map[string]string{"service": "mongodb", "action": "DescribeDBInstances"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "instance_name",
				Description: "The name of the MongoDB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_id",
				Description: "The ID of the MongoDB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_status",
				Description: "The status of the MongoDB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_type",
				Description: "The type of the MongoDB instance (ReplicaSet/ShardedCluster).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "db_engine",
				Description: "The database engine of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "db_engine_version",
				Description: "The database engine version of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "db_engine_version_str",
				Description: "The human-readable database engine version string.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "charge_type",
				Description: "The charge type of the MongoDB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "charge_status",
				Description: "The charge status of the MongoDB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "auto_renew",
				Description: "Whether auto-renewal is enabled for the instance.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC to which the instance belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subnet_id",
				Description: "The ID of the subnet to which the instance belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "zone_id",
				Description: "The ID of the zone to which the instance belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the instance belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "mongos_id",
				Description: "The Mongos ID of the sharded cluster instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "config_servers_id",
				Description: "The Config Servers ID of the sharded cluster instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "create_time",
				Description: "The time when the instance was created.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "update_time",
				Description: "The time when the instance was last updated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "expired_time",
				Description: "The expiration time of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "closed_time",
				Description: "The time when the instance was closed.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "reclaim_time",
				Description: "The time when the instance will be reclaimed.",
				Type:        proto.ColumnType_STRING,
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InstanceName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Hydrate:     getMongoDBInstanceARN,
				Transform:   transform.FromValue().Transform(transform.EnsureStringArray),
			},

			// Volcengine standard columns
			{
				Name:        "region",
				Description: ColumnDescriptionRegion,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromQual(matrixKeyRegion),
			},
			{
				Name:        "account_id",
				Description: ColumnDescriptionAccount,
				Type:        proto.ColumnType_STRING,
				Hydrate:     getCommonColumns,
				Transform:   transform.FromField("AccountID"),
			},
		},
	}
}

//// LIST FUNCTION

func listMongoDBInstances(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := MongoDBService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_mongodb_instance.listMongoDBInstances", "connection_error", err)
		return nil, err
	}

	input := &mongodb.DescribeDBInstancesInput{
		PageNumber: volcengine.Int32(1),
		PageSize:   volcengine.Int32(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "instance_name"); ok && value != nil {
		input.InstanceName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "instance_status"); ok && value != nil {
		input.InstanceStatus = value
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "zone_id"); ok && value != nil {
		input.ZoneId = value
	}

	if d.QueryContext.Limit != nil {
		limit := *d.QueryContext.Limit
		if limit < 100 {
			input.PageSize = volcengine.Int32(int32(limit))
		}
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.DescribeDBInstances(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_mongodb_instance.listMongoDBInstances", "query_error", err)
			return nil, err
		}

		if response.DBInstances == nil {
			break
		}

		for _, instance := range response.DBInstances {
			d.StreamListItem(ctx, instance)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		totalCount := int32(0)
		if response.Total != nil {
			totalCount = *response.Total
		}
		currentPage := int32(1)
		if input.PageNumber != nil {
			currentPage = *input.PageNumber
		}
		pageSize := int32(100)
		if input.PageSize != nil {
			pageSize = *input.PageSize
		}
		if currentPage*pageSize >= totalCount {
			break
		}
		input.PageNumber = volcengine.Int32(currentPage + 1)
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getMongoDBInstance(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := MongoDBService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_mongodb_instance.getMongoDBInstance", "connection_error", err)
		return nil, err
	}

	id := d.EqualsQuals["instance_id"].GetStringValue()
	if id == "" {
		return nil, nil
	}

	input := &mongodb.DescribeDBInstancesInput{
		InstanceId: volcengine.String(id),
		PageNumber: volcengine.Int32(1),
		PageSize:   volcengine.Int32(100),
	}

	response, err := client.DescribeDBInstances(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_mongodb_instance.getMongoDBInstance", "query_error", err)
		return nil, err
	}

	if response.DBInstances != nil && len(response.DBInstances) > 0 {
		return response.DBInstances[0], nil
	}

	return nil, nil
}

func getMongoDBInstanceARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	instance := h.Item.(*mongodb.DBInstanceForDescribeDBInstancesOutput)
	arn := "arn:volcengine:mongodb:" + region + "::instance/" + *instance.InstanceId
	return arn, nil
}
