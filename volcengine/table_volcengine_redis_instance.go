package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/redis"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineRedisInstance(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_redis_instance",
		Description: "Volcengine Redis Instance",
		List: &plugin.ListConfig{
			Hydrate: listRedisInstances,
			Tags:    map[string]string{"service": "redis", "action": "DescribeDBInstances"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("instance_id"),
			Hydrate:    getRedisInstance,
			Tags:       map[string]string{"service": "redis", "action": "DescribeDBInstances"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "instance_name",
				Description: "The name of the Redis instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_id",
				Description: "The ID of the Redis instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the Redis instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "engine_version",
				Description: "The engine version of the Redis instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_class",
				Description: "The instance class/specification of the Redis instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "charge_type",
				Description: "The charge type of the Redis instance (PrePaid/PostPaid).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC to which the Redis instance belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "region_id",
				Description: "The region ID of the Redis instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the Redis instance belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "multi_az",
				Description: "Whether the Redis instance is deployed across multiple availability zones.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "node_number",
				Description: "The number of nodes in the Redis instance.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "shard_number",
				Description: "The number of shards in the Redis instance.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "shard_capacity",
				Description: "The capacity of each shard in the Redis instance (in MB).",
				Type:        proto.ColumnType_DOUBLE,
			},
			{
				Name:        "sharded_cluster",
				Description: "Whether the Redis instance is a sharded cluster.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "zone_ids",
				Description: "The list of zone IDs where the Redis instance is deployed.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "capacity",
				Description: "The capacity usage of the Redis instance (total and used, in bytes).",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "create_time",
				Description: "The time when the Redis instance was created.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "expired_time",
				Description: "The expiration time of the Redis instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the Redis instance.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(redisTagsToMap),
			},
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
				Hydrate:     getRedisInstanceARN,
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

func listRedisInstances(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := RedisService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_redis_instance.listRedisInstances", "connection_error", err)
		return nil, err
	}

	region := d.EqualsQualString(matrixKeyRegion)

	input := &redis.DescribeDBInstancesInput{
		RegionId:   volcengine.String(region),
		PageNumber: volcengine.Int32(1),
		PageSize:   volcengine.Int32(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "instance_name"); ok && value != nil {
		input.InstanceName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "status"); ok && value != nil {
		input.Status = value
	}
	if value, ok := GetStringQualValue(d.Quals, "project_name"); ok && value != nil {
		input.ProjectName = value
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
			plugin.Logger(ctx).Error("volcengine_redis_instance.listRedisInstances", "query_error", err)
			return nil, err
		}

		if response.Instances == nil {
			break
		}

		for _, instance := range response.Instances {
			d.StreamListItem(ctx, instance)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		totalCount := int32(0)
		if response.TotalInstancesNum != nil {
			totalCount = *response.TotalInstancesNum
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

func getRedisInstance(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := RedisService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_redis_instance.getRedisInstance", "connection_error", err)
		return nil, err
	}

	region := d.EqualsQualString(matrixKeyRegion)

	id := d.EqualsQuals["instance_id"].GetStringValue()
	if id == "" {
		return nil, nil
	}

	input := &redis.DescribeDBInstancesInput{
		RegionId:   volcengine.String(region),
		InstanceId: volcengine.String(id),
		PageNumber: volcengine.Int32(1),
		PageSize:   volcengine.Int32(100),
	}

	response, err := client.DescribeDBInstances(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_redis_instance.getRedisInstance", "query_error", err)
		return nil, err
	}

	if response.Instances != nil && len(response.Instances) > 0 {
		return response.Instances[0], nil
	}

	return nil, nil
}

func getRedisInstanceARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	instance := h.Item.(*redis.InstanceForDescribeDBInstancesOutput)
	arn := "arn:volcengine:redis:" + region + "::instance/" + *instance.InstanceId
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func redisTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	switch tags := d.Value.(type) {
	case []*redis.TagForDescribeDBInstancesOutput:
		tagMap := map[string]string{}
		for _, tag := range tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}
		return tagMap, nil
	default:
		return nil, nil
	}
}
