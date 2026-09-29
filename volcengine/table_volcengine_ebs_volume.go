package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/storageebs"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineEbsVolume(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_ebs_volume",
		Description: "Volcengine Elastic Block Storage (EBS) Volume",
		List: &plugin.ListConfig{
			Hydrate: listEbsVolumes,
			Tags:    map[string]string{"service": "storageebs", "action": "DescribeVolumes"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("volume_id"),
			Hydrate:    getEbsVolume,
			Tags:       map[string]string{"service": "storageebs", "action": "DescribeVolumes"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "volume_name",
				Description: "The name of the volume.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "volume_id",
				Description: "The ID of the volume.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the volume.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "volume_type",
				Description: "The type of the volume (ESSD_PL0, ESSD_PL1, ESSD_PL2, ESSD_PL3, ESSD_FlexPL, PTSSD).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "size",
				Description: "The size of the volume (GiB).",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "description",
				Description: "The description of the volume.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "zone_id",
				Description: "The zone in which the volume resides.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instance_id",
				Description: "The ID of the instance to which the volume is attached.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "kind",
				Description: "The kind of the volume (system/data).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "charge_type",
				Description: "The billing method of the volume (PostPaid/PrePaid).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "pay_type",
				Description: "The pay type of the volume.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "encrypted",
				Description: "Indicates whether the volume is encrypted.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "image_id",
				Description: "The ID of the image used to create the volume.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "snapshot_id",
				Description: "The ID of the snapshot used to create the volume.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_at",
				Description: "The time when the volume was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "updated_at",
				Description: "The time when the volume was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "expired_at",
				Description: "The expiration time of the volume.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "iops",
				Description: "The IOPS of the volume.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Iops"),
			},
			{
				Name:        "throughput",
				Description: "The throughput of the volume (MiB/s).",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "delete_with_instance",
				Description: "Indicates whether the volume is released when its associated instance is released.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the volume belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the volume.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(ebsTagsToMap),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(ebsVolumeTitle),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Hydrate:     getEbsVolumeARN,
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

func listEbsVolumes(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := StorageEBSService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_ebs_volume.listEbsVolumes", "connection_error", err)
		return nil, err
	}

	input := &storageebs.DescribeVolumesInput{
		PageNumber: volcengine.Int32(1),
		PageSize:   volcengine.Int32(100),
	}

	// Apply optional filters
	if value, ok := GetStringQualValue(d.Quals, "volume_name"); ok && value != nil {
		input.VolumeName = value
	}
	if value, ok := GetStringQualValue(d.Quals, "status"); ok && value != nil {
		input.VolumeStatus = value
	}
	if value, ok := GetStringQualValue(d.Quals, "volume_type"); ok && value != nil {
		input.VolumeType = value
	}
	if value, ok := GetStringQualValue(d.Quals, "instance_id"); ok && value != nil {
		input.InstanceId = value
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
		response, err := client.DescribeVolumes(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_ebs_volume.listEbsVolumes", "query_error", err)
			return nil, err
		}

		if response.Volumes == nil {
			break
		}

		for _, volume := range response.Volumes {
			d.StreamListItem(ctx, volume)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		totalCount := int32(0)
		if response.TotalCount != nil {
			totalCount = *response.TotalCount
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

func getEbsVolume(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := StorageEBSService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_ebs_volume.getEbsVolume", "connection_error", err)
		return nil, err
	}

	var id string
	if h.Item != nil {
		volume := h.Item.(*storageebs.VolumeForDescribeVolumesOutput)
		id = *volume.VolumeId
	} else {
		id = d.EqualsQuals["volume_id"].GetStringValue()
	}

	if id == "" {
		return nil, nil
	}

	input := &storageebs.DescribeVolumesInput{
		VolumeIds: []*string{volcengine.String(id)},
	}

	response, err := client.DescribeVolumes(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_ebs_volume.getEbsVolume", "query_error", err)
		return nil, err
	}

	if response.Volumes != nil && len(response.Volumes) > 0 {
		return response.Volumes[0], nil
	}

	return nil, nil
}

func getEbsVolumeARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	volume := h.Item.(*storageebs.VolumeForDescribeVolumesOutput)
	region := d.EqualsQualString(matrixKeyRegion)

	arn := "arn:volcengine:storageebs:" + region + "::volume/" + *volume.VolumeId
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func ebsTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	tags, ok := d.Value.([]*storageebs.TagForDescribeVolumesOutput)
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

func ebsVolumeTitle(_ context.Context, d *transform.TransformData) (interface{}, error) {
	volume, ok := d.HydrateItem.(*storageebs.VolumeForDescribeVolumesOutput)
	if !ok {
		return nil, nil
	}

	if volume.VolumeName != nil && *volume.VolumeName != "" {
		return *volume.VolumeName, nil
	}

	return *volume.VolumeId, nil
}
