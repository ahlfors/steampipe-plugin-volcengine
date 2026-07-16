package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/alb"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableVolcengineAlb(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_alb",
		Description: "Volcengine Application Load Balancer (ALB)",
		List: &plugin.ListConfig{
			Hydrate: listAlbs,
			Tags:    map[string]string{"service": "alb", "action": "DescribeLoadBalancers"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("load_balancer_id"),
			Hydrate:    getAlb,
			Tags:       map[string]string{"service": "alb", "action": "DescribeLoadBalancers"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "load_balancer_name",
				Description: "The name of the ALB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "load_balancer_id",
				Description: "The ID of the ALB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the ALB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the ALB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The type of the ALB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "address_ip_version",
				Description: "The IP version of the ALB address.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "business_status",
				Description: "The business status of the ALB.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "dns_name",
				Description: "The DNS name of the ALB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "delete_protection",
				Description: "Whether delete protection is enabled.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "eip_address",
				Description: "The EIP address of the ALB instance.",
				Type:        proto.ColumnType_IPADDR,
			},
			{
				Name:        "eip_id",
				Description: "The EIP allocation ID.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "eni_address",
				Description: "The ENI address of the ALB instance.",
				Type:        proto.ColumnType_IPADDR,
			},
			{
				Name:        "eni_id",
				Description: "The ENI ID of the ALB instance.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC to which the ALB belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subnet_id",
				Description: "The ID of the subnet to which the ALB belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "load_balancer_billing_type",
				Description: "The billing type of the ALB instance.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "lock_reason",
				Description: "The reason the ALB is locked.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the ALB belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "local_addresses",
				Description: "The local addresses of the ALB instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "zone_mappings",
				Description: "The zone mappings of the ALB instance.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "create_time",
				Description: "The time when the ALB was created.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "update_time",
				Description: "The time when the ALB was last updated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "deleted_time",
				Description: "The time when the ALB was deleted.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "overdue_time",
				Description: "The overdue time of the ALB.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the ALB.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(albTagsToMap),
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
				Hydrate:     getAlbARN,
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

func listAlbs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := ALBService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_alb.listAlbs", "connection_error", err)
		return nil, err
	}

	input := &alb.DescribeLoadBalancersInput{
		PageNumber: volcengine.Int64(1),
		PageSize:   volcengine.Int64(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "load_balancer_name"); ok && value != nil {
		input.LoadBalancerName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "project_name"); ok && value != nil {
		input.ProjectName = value
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
			plugin.Logger(ctx).Error("volcengine_alb.listAlbs", "query_error", err)
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

func getAlb(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := ALBService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_alb.getAlb", "connection_error", err)
		return nil, err
	}

	id := d.EqualsQuals["load_balancer_id"].GetStringValue()
	if id == "" {
		return nil, nil
	}

	input := &alb.DescribeLoadBalancersInput{
		LoadBalancerIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeLoadBalancers(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_alb.getAlb", "query_error", err)
		return nil, err
	}

	if response.LoadBalancers != nil && len(response.LoadBalancers) > 0 {
		return response.LoadBalancers[0], nil
	}

	return nil, nil
}

func getAlbARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	lb := h.Item.(*alb.LoadBalancerForDescribeLoadBalancersOutput)
	arn := "arn:volcengine:alb:" + region + "::loadbalancer/" + *lb.LoadBalancerId
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func albTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	switch tags := d.Value.(type) {
	case []*alb.TagForDescribeLoadBalancersOutput:
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
