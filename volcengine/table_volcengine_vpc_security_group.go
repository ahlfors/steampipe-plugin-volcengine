package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/vpc"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableVolcengineVpcSecurityGroup(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_vpc_security_group",
		Description: "Volcengine VPC Security Group",
		List: &plugin.ListConfig{
			Hydrate: listVpcSecurityGroups,
			Tags:    map[string]string{"service": "vpc", "action": "DescribeSecurityGroups"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("security_group_id"),
			Hydrate:    getVpcSecurityGroup,
			Tags:       map[string]string{"service": "vpc", "action": "DescribeSecurityGroups"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "security_group_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the security group.",
			},
			{
				Name:        "security_group_name",
				Type:        proto.ColumnType_STRING,
				Description: "The name of the security group.",
			},
			{
				Name:        "vpc_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the VPC to which the security group belongs.",
			},
			{
				Name:        "description",
				Type:        proto.ColumnType_STRING,
				Description: "The description of the security group.",
			},
			{
				Name:        "status",
				Type:        proto.ColumnType_STRING,
				Description: "The status of the security group.",
			},
			{
				Name:        "created_at",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the security group was created.",
			},
			{
				Name:        "security_group_type",
				Type:        proto.ColumnType_STRING,
				Description: "The type of the security group (default/custom).",
			},
			{
				Name:        "project_name",
				Type:        proto.ColumnType_STRING,
				Description: "The name of the project.",
			},
			{
				Name:        "tags_src",
				Type:        proto.ColumnType_JSON,
				Description: "A list of tags assigned to the security group.",
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
				Transform:   transform.FromField("SecurityGroupName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("SecurityGroupId").Transform(ensureStringArray),
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

func listVpcSecurityGroups(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc_security_group.listVpcSecurityGroups", "connection_error", err)
		return nil, err
	}

	input := &vpc.DescribeSecurityGroupsInput{
		MaxResults: volcengine.Int32(100),
	}

	if value, ok := GetStringQualValue(d.Quals, "security_group_id"); ok && value != nil {
		input.SecurityGroupIds = append(input.SecurityGroupIds, value)
	}
	if value, ok := GetStringQualValue(d.Quals, "vpc_id"); ok && value != nil {
		input.VpcId = value
	}
	if value, ok := GetStringQualValue(d.Quals, "security_group_name"); ok && value != nil {
		input.SecurityGroupName = value
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.DescribeSecurityGroups(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_vpc_security_group.listVpcSecurityGroups", "query_error", err)
			return nil, err
		}

		if response.SecurityGroups == nil {
			break
		}

		for _, sg := range response.SecurityGroups {
			d.StreamListItem(ctx, sg)
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

func getVpcSecurityGroup(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc_security_group.getVpcSecurityGroup", "connection_error", err)
		return nil, err
	}

	var id string
	if h.Item != nil {
		sg := h.Item.(*vpc.SecurityGroupForDescribeSecurityGroupsOutput)
		id = *sg.SecurityGroupId
	} else {
		id = d.EqualsQuals["security_group_id"].GetStringValue()
	}

	if id == "" {
		return nil, nil
	}

	input := &vpc.DescribeSecurityGroupsInput{
		SecurityGroupIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeSecurityGroups(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vpc_security_group.getVpcSecurityGroup", "query_error", err)
		return nil, err
	}

	if response.SecurityGroups != nil && len(response.SecurityGroups) > 0 {
		return response.SecurityGroups[0], nil
	}
	return nil, nil
}
