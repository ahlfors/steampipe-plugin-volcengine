package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/vpc"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableVolcengineVpcSubnet(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_vpc_subnet",
		Description: "Volcengine VPC Subnet",
		List: &plugin.ListConfig{
			Hydrate: listVpcSubnets,
			Tags:    map[string]string{"service": "vpc", "action": "DescribeSubnets"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("subnet_id"),
			Hydrate:    getVpcSubnet,
			Tags:       map[string]string{"service": "vpc", "action": "DescribeSubnets"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "subnet_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the subnet.",
			},
			{
				Name:        "subnet_name",
				Type:        proto.ColumnType_STRING,
				Description: "The name of the subnet.",
			},
			{
				Name:        "vpc_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the VPC to which the subnet belongs.",
			},
			{
				Name:        "zone_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the availability zone.",
			},
			{
				Name:        "cidr_block",
				Type:        proto.ColumnType_CIDR,
				Description: "The CIDR block of the subnet.",
			},
			{
				Name:        "status",
				Type:        proto.ColumnType_STRING,
				Description: "The status of the subnet.",
			},
			{
				Name:        "available_ip_address_count",
				Type:        proto.ColumnType_INT,
				Description: "The number of available IP addresses.",
			},
			{
				Name:        "total_ipv4_count",
				Type:        proto.ColumnType_INT,
				Description: "Total number of IPv4 addresses.",
			},
			{
				Name:        "description",
				Type:        proto.ColumnType_STRING,
				Description: "The description of the subnet.",
			},
			{
				Name:        "route_table_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the route table.",
			},
			{
				Name:        "created_at",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the subnet was created.",
			},
			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("SubnetName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("SubnetId").Transform(ensureStringArray),
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

func listVpcSubnets(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc_subnet.listVpcSubnets", "connection_error", err)
		return nil, err
	}

	input := &vpc.DescribeSubnetsInput{
		MaxResults: volcengine.Int32(100),
	}

	if value, ok := GetStringQualValue(d.Quals, "subnet_id"); ok && value != nil {
		input.SubnetIds = append(input.SubnetIds, value)
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "subnet_name"); ok && value != nil {
		input.SubnetName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "zone_id"); ok && value != nil {
		input.ZoneId = value
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.DescribeSubnets(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_vpc_subnet.listVpcSubnets", "query_error", err)
			return nil, err
		}

		if response.Subnets == nil {
			break
		}

		for _, subnet := range response.Subnets {
			d.StreamListItem(ctx, subnet)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if response.NextToken == nil || *response.NextToken == "" {
			break
		}
		input.NextToken = response.NextToken
	}

	return nil, nil
}

//// GET FUNCTION

func getVpcSubnet(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc_subnet.getVpcSubnet", "connection_error", err)
		return nil, err
	}

	var id string
	if h.Item != nil {
		subnet := h.Item.(*vpc.SubnetForDescribeSubnetsOutput)
		id = *subnet.SubnetId
	} else {
		id = d.EqualsQuals["subnet_id"].GetStringValue()
	}

	if id == "" {
		return nil, nil
	}

	input := &vpc.DescribeSubnetsInput{
		SubnetIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeSubnets(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc_subnet.getVpcSubnet", "query_error", err)
		return nil, err
	}

	if response.Subnets != nil && len(response.Subnets) > 0 {
		return response.Subnets[0], nil
	}
	return nil, nil
}
