package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/vke"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineVkeCluster(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_vke_cluster",
		Description: "Volcengine Kubernetes Engine (VKE) Cluster",
		List: &plugin.ListConfig{
			Hydrate: listVkeClusters,
			Tags:    map[string]string{"service": "vke", "action": "ListClusters"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getVkeCluster,
			Tags:       map[string]string{"service": "vke", "action": "ListClusters"},
		},
		GetMatrixItemFunc: BuildRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the VKE cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The ID of the VKE cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "kubernetes_version",
				Description: "The Kubernetes version of the cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the VKE cluster.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status_phase",
				Description: "The current phase of the cluster (Creating/Running/Updating/Deleting/Failed/Starting/Stopped).",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status.Phase"),
			},
			{
				Name:        "status_conditions",
				Description: "The status conditions of the cluster.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Status.Conditions"),
			},
			{
				Name:        "delete_protection_enabled",
				Description: "Whether delete protection is enabled for the cluster.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "cluster_config",
				Description: "The cluster configuration including VPC, subnets, security groups, and API server endpoints.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "vpc_id",
				Description: "The ID of the VPC to which the cluster belongs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ClusterConfig.VpcId"),
			},
			{
				Name:        "pods_config",
				Description: "The pods network configuration (Flannel/VpcCni/Calico).",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "services_config",
				Description: "The services CIDR configuration.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "logging_config",
				Description: "The logging configuration of the cluster.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "node_statistics",
				Description: "The node statistics (total, running, creating, deleting, failed, updating counts).",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "node_total_count",
				Description: "The total number of nodes in the cluster.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("NodeStatistics.TotalCount"),
			},
			{
				Name:        "node_running_count",
				Description: "The number of running nodes in the cluster.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("NodeStatistics.RunningCount"),
			},
			{
				Name:        "create_time",
				Description: "The time when the cluster was created.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "update_time",
				Description: "The time when the cluster was last updated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the cluster.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Steampipe standard columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags").Transform(vkeTagsToMap),
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
				Hydrate:     getVkeClusterARN,
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

func listVkeClusters(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := VKEService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vke_cluster.listVkeClusters", "connection_error", err)
		return nil, err
	}

	input := &vke.ListClustersInput{
		PageNumber: volcengine.Int32(1),
		PageSize:   volcengine.Int32(100),
	}

	// Build filter from quals
	filter := &vke.FilterForListClustersInput{}
	hasFilter := false

	if value, ok := GetStringQualValue(d.Quals, "name"); ok && value != nil {
		filter.Name = value
		hasFilter = true
	}

	if hasFilter {
		input.Filter = filter
	}

	if d.QueryContext.Limit != nil {
		limit := *d.QueryContext.Limit
		if limit < 100 {
			input.PageSize = volcengine.Int32(int32(limit))
		}
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.ListClusters(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_vke_cluster.listVkeClusters", "query_error", err)
			return nil, err
		}

		if response.Items == nil {
			break
		}

		for _, cluster := range response.Items {
			d.StreamListItem(ctx, cluster)
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

func getVkeCluster(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := VKEService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vke_cluster.getVkeCluster", "connection_error", err)
		return nil, err
	}

	id := d.EqualsQuals["id"].GetStringValue()
	if id == "" {
		return nil, nil
	}

	input := &vke.ListClustersInput{
		PageNumber: volcengine.Int32(1),
		PageSize:   volcengine.Int32(100),
		Filter: &vke.FilterForListClustersInput{
			Ids: []*string{volcengine.String(id)},
		},
	}

	response, err := client.ListClusters(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_vke_cluster.getVkeCluster", "query_error", err)
		return nil, err
	}

	if response.Items != nil && len(response.Items) > 0 {
		return response.Items[0], nil
	}

	return nil, nil
}

func getVkeClusterARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	cluster := h.Item.(*vke.ItemForListClustersOutput)
	arn := "arn:volcengine:vke:" + region + "::cluster/" + *cluster.Id
	return arn, nil
}

//// TRANSFORM FUNCTIONS

func vkeTagsToMap(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}

	switch tags := d.Value.(type) {
	case []*vke.TagForListClustersOutput:
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
