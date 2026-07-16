package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/ecs"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableVolcengineEcsInstance(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_ecs_instance",
		Description: "Volcengine Elastic Compute Service (ECS) Instance",
		List: &plugin.ListConfig{
			Hydrate: listEcsInstances,
			Tags:    map[string]string{"service": "ecs", "action": "DescribeInstances"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("instance_id"),
			Hydrate:    getEcsInstance,
			Tags:       map[string]string{"service": "ecs", "action": "DescribeInstances"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "instance_name",
				Description: "The name of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_id",
				Description: "The ID of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_type_id",
				Description: "The instance type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the instance.",
				Type:        proto.ColumnType_STRING,
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
				Description: "The zone in which the instance resides.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "image_id",
				Description: "The ID of the image used to create the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "os_name",
				Description: "The name of the operating system.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OsName"),
			},
			{
				Name:        "os_type",
				Description: "The type of the operating system (linux/windows).",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("OsType"),
			},
			{
				Name:        "cpu",
				Description: "The number of vCPUs.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "memory_size",
				Description: "The memory size of the instance (in MiB).",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "created_at",
				Description: "The time when the instance was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the instance was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "expired_at",
				Description: "The expiration time of the instance.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "instance_charge_type",
				Description: "The billing method of the instance (PostPaid/PrePaid).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "key_pair_name",
				Description: "The name of the SSH key pair.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "host_name",
				Description: "The hostname of the instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "security_group_ids",
				Description: "The IDs of the security groups to which the instance belongs.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "network_interfaces",
				Description: "The network interfaces of the instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the instance.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(ecsTagsToMap),
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
				Hydrate:     getEcsInstanceARN,
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

func listEcsInstances(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := ECSService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_ecs_instance.listEcsInstances", "connection_error", err)
		return nil, err
	}

	_ = d.EqualsQualString(matrixKeyRegion)
	input := &ecs.DescribeInstancesInput{
		MaxResults: volcengine.Int32(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "instance_name"); ok && value != nil {
		input.InstanceName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "status"); ok && value != nil {
		input.Status = value
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "zone_id"); ok && value != nil {
		input.ZoneId = value
	}

	// Adjust page size if limit is specified
	if d.QueryContext.Limit != nil {
		limit := *d.QueryContext.Limit
		if limit < 100 {
			input.MaxResults = volcengine.Int32(int32(limit))
		}
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.DescribeInstances(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_ecs_instance.listEcsInstances", "query_error", err)
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

		// Check if there are more pages using NextToken
		if response.NextToken == nil || *response.NextToken == "" {
			break
		}
		input.NextToken = response.NextToken
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getEcsInstance(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := ECSService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_ecs_instance.getEcsInstance", "connection_error", err)
		return nil, err
	}

	var id string
	if h.Item != nil {
		instance := h.Item.(*ecs.InstanceForDescribeInstancesOutput)
		id = *instance.InstanceId
	} else {
		id = d.EqualsQuals["instance_id"].GetStringValue()
	}

	if id == "" {
		return nil, nil
	}

	input := &ecs.DescribeInstancesInput{
		InstanceIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeInstances(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_ecs_instance.getEcsInstance", "query_error", err)
		return nil, err
	}

	if response.Instances != nil && len(response.Instances) > 0 {
		return response.Instances[0], nil
	}

	return nil, nil
}

func getEcsInstanceARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	instance := h.Item.(*ecs.InstanceForDescribeInstancesOutput)
	region := d.EqualsQualString(matrixKeyRegion)

	// Volcengine ARN format: arn:volcengine:ecs:{region}:{account_id}:instance/{instance_id}
	arn := "arn:volcengine:ecs:" + region + "::instance/" + *instance.InstanceId
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func ecsTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	tags, ok := d.Value.([]*ecs.TagForDescribeInstancesOutput)
	if !ok || len(tags) == 0 {
		return nil, nil
	}

	tagMap := map[string]string{}
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			tagMap[*tag.Key] = *tag.Value
		}
	}

	return tagMap, nil
}
