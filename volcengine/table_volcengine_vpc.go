package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/vpc"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineVpc(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_vpc",
		Description: "Volcengine Virtual Private Cloud (VPC)",
		List: &plugin.ListConfig{
			Hydrate: listVpcs,
			Tags:    map[string]string{"service": "vpc", "action": "DescribeVpcs"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("vpc_id"),
			Hydrate:    getVpc,
			Tags:       map[string]string{"service": "vpc", "action": "DescribeVpcs"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "vpc_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the VPC.",
			},
			{
				Name:        "vpc_name",
				Type:        proto.ColumnType_STRING,
				Description: "The name of the VPC.",
			},
			{
				Name:        "cidr_block",
				Type:        proto.ColumnType_CIDR,
				Description: "The CIDR block of the VPC.",
			},
			{
				Name:        "status",
				Type:        proto.ColumnType_STRING,
				Description: "The status of the VPC (Available, Pending).",
			},
			{
				Name:        "enable_ipv6",
				Type:        proto.ColumnType_BOOL,
				Description: "Whether IPv6 is enabled.",
			},
			{
				Name:        "description",
				Type:        proto.ColumnType_STRING,
				Description: "The description of the VPC.",
			},
			{
				Name:        "created_at",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the VPC was created.",
			},
			{
				Name:        "updated_at",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the VPC was last updated.",
			},
			{
				Name:        "tags_src",
				Type:        proto.ColumnType_JSON,
				Description: "A list of tags assigned to the VPC.",
				Transform:   transform.FromField("Tags"),
			},
			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(vpcTagsToMap),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("VpcName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("VpcId").Transform(ensureStringArray),
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

func listVpcs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc.listVpcs", "connection_error", err)
		return nil, err
	}

	input := &vpc.DescribeVpcsInput{
		MaxResults: volcengine.Int64(100),
	}

	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcIds = append(input.VpcIds, value)
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_name"); ok && value != nil {
		input.VpcName = value
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.DescribeVpcs(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_vpc.listVpcs", "query_error", err)
			return nil, err
		}

		if response.Vpcs == nil {
			break
		}

		for _, vpc := range response.Vpcs {
			d.StreamListItem(ctx, vpc)
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

func getVpc(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc.getVpc", "connection_error", err)
		return nil, err
	}

	var id string
	if h.Item != nil {
		vpc := h.Item.(*vpc.VpcForDescribeVpcsOutput)
		id = *vpc.VpcId
	} else {
		id = d.EqualsQuals["vpc_id"].GetStringValue()
	}

	if id == "" {
		return nil, nil
	}

	input := &vpc.DescribeVpcsInput{
		VpcIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeVpcs(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc.getVpc", "query_error", err)
		return nil, err
	}

	if response.Vpcs != nil && len(response.Vpcs) > 0 {
		return response.Vpcs[0], nil
	}
	return nil, nil
}

//// TRANSFORM FUNCTIONS

func vpcTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	tags, ok := d.Value.([]*vpc.TagForDescribeVpcsOutput)
	if !ok || len(tags) == 0 {
		return nil, nil
	}

	result := map[string]string{}
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
		}
	}
	return result, nil
}
