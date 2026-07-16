package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/clb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableVolcengineClb(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_clb",
		Description: "Volcengine Cloud Load Balancer (CLB)",
		List: &plugin.ListConfig{
			Hydrate: listClbs,
			Tags:    map[string]string{"service": "clb", "action": "DescribeLoadBalancers"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("load_balancer_id"),
			Hydrate:    getClb,
			Tags:       map[string]string{"service": "clb", "action": "DescribeLoadBalancers"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "load_balancer_name",
				Description: "The name of the CLB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "load_balancer_id",
				Description: "The ID of the CLB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the CLB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the CLB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "address_type",
				Description: "The address type of the CLB (public/private).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The type of the CLB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "specification",
				Description: "The specification of the CLB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC to which the CLB belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subnet_id",
				Description: "The ID of the subnet to which the CLB belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "master_zone_id",
				Description: "The ID of the primary zone.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "slave_zone_id",
				Description: "The ID of the secondary zone.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "eip_address",
				Description: "The EIP address of the CLB instance.",
				Type:        proto.ColumnType_IPADDR,
			},
			{
				Name:        "eip_allocation_id",
				Description: "The EIP allocation ID.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "ipv6_address",
				Description: "The IPv6 address of the CLB instance.",
				Type:        proto.ColumnType_IPADDR,
			},
			{
				Name:        "private_ip_address",
				Description: "The private IP address of the CLB instance.",
				Type:        proto.ColumnType_IPADDR,
			},
			{
				Name:        "created_at",
				Description: "The time when the CLB was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the CLB was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "expired_at",
				Description: "The expiration time of the CLB.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "business_status",
				Description: "The business status of the CLB.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the CLB belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "server_group_id",
				Description: "The ID of the default server group.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the CLB.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(clbTagsToMap),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("LoadBalancerName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Hydrate:     getClbARN,
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

func listClbs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := CLBService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_clb.listClbs", "connection_error", err)
		return nil, err
	}

	input := &clb.DescribeLoadBalancersInput{
		PageNumber: volcengine.Int64(1),
		PageSize:   volcengine.Int64(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "load_balancer_name"); ok && value != nil {
		input.LoadBalancerName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "status"); ok && value != nil {
		input.Status = value
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}

	if d.QueryContext.Limit != nil {
		limit := *d.QueryContext.Limit
		if limit < 100 {
			input.PageSize = volcengine.Int64(limit)
		}
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.DescribeLoadBalancers(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_clb.listClbs", "query_error", err)
			return nil, err
		}

		if response.LoadBalancers == nil {
			break
		}

		for _, lb := range response.LoadBalancers {
			d.StreamListItem(ctx, lb)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		totalCount := int64(0)
		if response.TotalCount != nil {
			totalCount = *response.TotalCount
		}
		currentPage := int64(1)
		if input.PageNumber != nil {
			currentPage = *input.PageNumber
		}
		pageSize := int64(100)
		if input.PageSize != nil {
			pageSize = *input.PageSize
		}
		if currentPage*pageSize >= totalCount {
			break
		}
		input.PageNumber = volcengine.Int64(currentPage + 1)
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getClb(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := CLBService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_clb.getClb", "connection_error", err)
		return nil, err
	}

	id := d.EqualsQuals["load_balancer_id"].GetStringValue()
	if id == "" {
		return nil, nil
	}

	input := &clb.DescribeLoadBalancersInput{
		LoadBalancerIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeLoadBalancers(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_clb.getClb", "query_error", err)
		return nil, err
	}

	if response.LoadBalancers != nil && len(response.LoadBalancers) > 0 {
		return response.LoadBalancers[0], nil
	}

	return nil, nil
}

func getClbARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	lb := h.Item.(*clb.LoadBalancerForDescribeLoadBalancersOutput)
	arn := "arn:volcengine:clb:" + region + "::loadbalancer/" + *lb.LoadBalancerId
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func clbTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	// The exact type depends on the SDK struct
	// Handle common patterns
	switch tags := d.Value.(type) {
	case []*clb.TagForDescribeLoadBalancersOutput:
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
