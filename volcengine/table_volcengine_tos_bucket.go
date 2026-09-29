package volcengine

import (
	"context"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineTosBucket(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_tos_bucket",
		Description: "Volcengine TOS (Tinder Object Storage) Bucket",
		List: &plugin.ListConfig{
			Hydrate: listTosBuckets,
			Tags:    map[string]string{"service": "tos", "action": "ListBuckets"},
		},
		HydrateConfig: []plugin.HydrateConfig{
			{
				Func: getBucketInfo,
				Tags: map[string]string{"service": "tos", "action": "HeadBucket"},
			},
			{
				Func: getBucketTags,
				Tags: map[string]string{"service": "tos", "action": "GetBucketTagging"},
			},
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the bucket.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "location",
				Description: "The region where the bucket is located.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "creation_date",
				Description: "The date when the bucket was created.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CreationDate"),
			},
			{
				Name:        "extranet_endpoint",
				Description: "The extranet endpoint of the bucket.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "intranet_endpoint",
				Description: "The intranet endpoint of the bucket.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "The name of the project to which the bucket belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "storage_class",
				Description: "The default storage class of the bucket.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getBucketInfo,
				Transform:   transform.FromField("StorageClass"),
			},
			{
				Name:        "az_redundancy",
				Description: "The availability zone redundancy of the bucket.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getBucketInfo,
				Transform:   transform.FromField("AzRedundancy"),
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the bucket.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getBucketTags,
				Transform:   transform.FromField("TagSet.Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Hydrate:     getBucketTags,
				Transform:   transform.FromField("TagSet.Tags").Transform(tosBucketTagsToMap),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(tosBucketARN).Transform(transform.EnsureStringArray),
			},

			// Volcengine standard columns
			{
				Name:        "region",
				Description: ColumnDescriptionRegion,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Location"),
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

func listTosBuckets(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// TOS ListBuckets is a global operation, use default region
	region := GetDefaultRegion(d.Connection)
	client, err := TOSService(ctx, d, region)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_tos_bucket.listTosBuckets", "connection_error", err)
		return nil, err
	}

	d.WaitForListRateLimit(ctx)
	output, err := client.ListBuckets(ctx, &tos.ListBucketsInput{})
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_tos_bucket.listTosBuckets", "query_error", err)
		return nil, err
	}

	for _, bucket := range output.Buckets {
		d.StreamListItem(ctx, bucket)
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getBucketInfo(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	bucket := h.Item.(tos.ListedBucket)
	bucketName := bucket.Name
	bucketLocation := bucket.Location

	if bucketName == "" {
		return nil, nil
	}

	region := bucketLocation
	if region == "" {
		region = GetDefaultRegion(d.Connection)
	}

	client, err := TOSService(ctx, d, region)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_tos_bucket.getBucketInfo", "connection_error", err)
		return nil, err
	}

	output, err := client.HeadBucket(ctx, &tos.HeadBucketInput{
		Bucket: bucketName,
	})
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_tos_bucket.getBucketInfo", "query_error", err, "bucket", bucketName)
		return nil, err
	}

	return output, nil
}

func getBucketTags(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	bucket := h.Item.(tos.ListedBucket)
	bucketName := bucket.Name
	bucketLocation := bucket.Location

	if bucketName == "" {
		return nil, nil
	}

	region := bucketLocation
	if region == "" {
		region = GetDefaultRegion(d.Connection)
	}

	client, err := TOSService(ctx, d, region)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_tos_bucket.getBucketTags", "connection_error", err)
		return nil, err
	}

	output, err := client.GetBucketTagging(ctx, &tos.GetBucketTaggingInput{
		Bucket: bucketName,
	})
	if err != nil {
		// TOS returns "The TagSet does not exist" (EC=0018-00000103) when
		// a bucket has no tags. Treat this as a non-error case.
		if strings.Contains(err.Error(), "TagSet does not exist") || strings.Contains(err.Error(), "0018-00000103") {
			return nil, nil
		}
		plugin.Logger(ctx).Error("volcengine_tos_bucket.getBucketTags", "query_error", err, "bucket", bucketName)
		return nil, err
	}

	return output, nil
}

//// TRANSFORM FUNCTIONS

func tosBucketTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	tags, ok := d.Value.([]tos.Tag)
	if !ok || len(tags) == 0 {
		return nil, nil
	}

	tagMap := map[string]string{}
	for _, tag := range tags {
		if tag.Key != "" {
			tagMap[tag.Key] = tag.Value
		}
	}

	return tagMap, nil
}

func tosBucketARN(_ context.Context, d *transform.TransformData) (interface{}, error) {
	bucket := d.HydrateItem.(tos.ListedBucket)
	return "arn:volcengine:tos:::" + bucket.Name, nil
}
