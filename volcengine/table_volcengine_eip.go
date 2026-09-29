package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/vpc"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineEip(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_eip",
		Description: "Volcengine Elastic IP Address (EIP)",
		List: &plugin.ListConfig{
			Hydrate: listEips,
			Tags:    map[string]string{"service": "vpc", "action": "DescribeEipAddresses"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("allocation_id"),
			Hydrate:    getEip,
			Tags:       map[string]string{"service": "vpc", "action": "DescribeEipAddresses"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the EIP.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "allocation_id",
				Description: "The unique ID of the EIP.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "eip_address",
				Description: "The IP address of the EIP.",
				Type:        proto.ColumnType_IPADDR,
			},
			{
				Name:        "status",
				Description: "The status of the EIP.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "bandwidth",
				Description: "The peak bandwidth of the EIP (Mbit/s).",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "billing_type",
				Description: "The billing type of the EIP.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "internet_charge_type",
				Description: "The metering method of the EIP.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the EIP.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_id",
				Description: "The ID of the instance to which the EIP is bound.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_type",
				Description: "The type of the instance to which the EIP is bound.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_at",
				Description: "The time when the EIP was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the EIP was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "expired_at",
				Description: "The expiration time of the EIP.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the EIP belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the EIP.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(eipTagsToMap),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(eipTitle),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Hydrate:     getEipARN,
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

func listEips(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_eip.listEips", "connection_error", err)
		return nil, err
	}

	input := &vpc.DescribeEipAddressesInput{
		PageNumber: volcengine.Int64(1),
		PageSize:   volcengine.Int64(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "name"); ok && value != nil {
		input.Name = value
	}
	if value, ok := GetStringQualValue(d.Quals, "status"); ok && value != nil {
		input.Status = value
	}

	if d.QueryContext.Limit != nil {
		limit := *d.QueryContext.Limit
		if limit < 100 {
			input.PageSize = volcengine.Int64(limit)
		}
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.DescribeEipAddresses(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_eip.listEips", "query_error", err)
			return nil, err
		}

		if response.EipAddresses == nil {
			break
		}

		for _, eip := range response.EipAddresses {
			d.StreamListItem(ctx, eip)
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

func getEip(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := VPCService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_eip.getEip", "connection_error", err)
		return nil, err
	}

	var id string
	if h.Item != nil {
		eip := h.Item.(*vpc.EipAddressForDescribeEipAddressesOutput)
		id = *eip.AllocationId
	} else {
		id = d.EqualsQuals["allocation_id"].GetStringValue()
	}

	if id == "" {
		return nil, nil
	}

	input := &vpc.DescribeEipAddressesInput{
		AllocationIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeEipAddresses(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_eip.getEip", "query_error", err)
		return nil, err
	}

	if response.EipAddresses != nil && len(response.EipAddresses) > 0 {
		return response.EipAddresses[0], nil
	}
	return nil, nil
}

func getEipARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var eip *vpc.EipAddressForDescribeEipAddressesOutput
	switch item := h.Item.(type) {
	case *vpc.EipAddressForDescribeEipAddressesOutput:
		eip = item
	}
	region := d.EqualsQualString(matrixKeyRegion)
	arn := "arn:volcengine:vpc:" + region + "::eip/" + *eip.AllocationId
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func eipTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	tags, ok := d.Value.([]*vpc.TagForDescribeEipAddressesOutput)
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

func eipTitle(_ context.Context, d *transform.TransformData) (interface{}, error) {
	eip, ok := d.HydrateItem.(*vpc.EipAddressForDescribeEipAddressesOutput)
	if !ok {
		return nil, nil
	}

	if eip.Name != nil && *eip.Name != "" {
		return *eip.Name, nil
	}

	return *eip.AllocationId, nil
}
