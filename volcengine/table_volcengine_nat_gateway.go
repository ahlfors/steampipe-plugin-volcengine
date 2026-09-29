package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/natgateway"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineNatGateway(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_nat_gateway",
		Description: "Volcengine NAT Gateway",
		List: &plugin.ListConfig{
			Hydrate: listNatGateways,
			Tags:    map[string]string{"service": "natgateway", "action": "DescribeNatGateways"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("nat_gateway_id"),
			Hydrate:    getNatGateway,
			Tags:       map[string]string{"service": "natgateway", "action": "DescribeNatGateways"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "nat_gateway_name",
				Description: "The name of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "nat_gateway_id",
				Description: "The ID of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "spec",
				Description: "The specification of the NAT Gateway (Small/Medium/Large/Extra_Large_1/Extra_Large_2).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "network_type",
				Description: "The network type of the NAT Gateway (internet/intranet).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "business_status",
				Description: "The business status of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "billing_type",
				Description: "The billing type of the NAT Gateway.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC to which the NAT Gateway belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subnet_id",
				Description: "The ID of the subnet to which the NAT Gateway belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "zone_id",
				Description: "The ID of the zone to which the NAT Gateway belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "network_interface_id",
				Description: "The network interface ID of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "lock_reason",
				Description: "The reason the NAT Gateway is locked.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the NAT Gateway belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "eip_addresses",
				Description: "The EIP addresses associated with the NAT Gateway.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "snat_entry_ids",
				Description: "The SNAT entry IDs associated with the NAT Gateway.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "dnat_entry_ids",
				Description: "The DNAT entry IDs associated with the NAT Gateway.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "creation_time",
				Description: "The time when the NAT Gateway was created.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "updated_at",
				Description: "The time when the NAT Gateway was last updated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "expired_time",
				Description: "The expiration time of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "deleted_time",
				Description: "The time when the NAT Gateway was deleted.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "overdue_time",
				Description: "The overdue time of the NAT Gateway.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the NAT Gateway.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(natGatewayTagsToMap),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("NatGatewayName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Hydrate:     getNatGatewayARN,
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

func listNatGateways(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := NatGatewayService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_nat_gateway.listNatGateways", "connection_error", err)
		return nil, err
	}

	input := &natgateway.DescribeNatGatewaysInput{
		PageNumber: volcengine.Int64(1),
		PageSize:   volcengine.Int64(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "nat_gateway_name"); ok && value != nil {
		input.NatGatewayName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "subnet_id"); ok && value != nil {
		input.SubnetId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "spec"); ok && value != nil {
		input.Spec = value
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
		response, err := client.DescribeNatGateways(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_nat_gateway.listNatGateways", "query_error", err)
			return nil, err
		}

		if response.NatGateways == nil {
			break
		}

		for _, ngw := range response.NatGateways {
			d.StreamListItem(ctx, ngw)
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

func getNatGateway(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := NatGatewayService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_nat_gateway.getNatGateway", "connection_error", err)
		return nil, err
	}

	id := d.EqualsQuals["nat_gateway_id"].GetStringValue()
	if id == "" {
		return nil, nil
	}

	input := &natgateway.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeNatGateways(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_nat_gateway.getNatGateway", "query_error", err)
		return nil, err
	}

	if response.NatGateways != nil && len(response.NatGateways) > 0 {
		return response.NatGateways[0], nil
	}

	return nil, nil
}

func getNatGatewayARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	ngw := h.Item.(*natgateway.NatGatewayForDescribeNatGatewaysOutput)
	arn := "arn:volcengine:natgateway:" + region + "::natgateway/" + *ngw.NatGatewayId
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func natGatewayTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	switch tags := d.Value.(type) {
	case []*natgateway.TagForDescribeNatGatewaysOutput:
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
