package volcengine

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// Plugin creates this (volcengine) plugin.
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             "steampipe-plugin-volcengine",
		DefaultTransform: transform.FromCamel().NullIfZero(),
		DefaultGetConfig: &plugin.GetConfig{
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: isNotFoundError([]string{"ResourceNotFound", "NotFound", "InvalidResource.NotFound"}),
			},
		},
		DefaultIgnoreConfig: &plugin.IgnoreConfig{
			ShouldIgnoreErrorFunc: shouldIgnoreErrorPluginDefault(),
		},
		ConnectionKeyColumns: []plugin.ConnectionKeyColumn{
			{
				Name:    "account_id",
				Hydrate: getAccountId,
			},
		},
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
		},
		TableMap: map[string]*plugin.Table{
			// Compute
			"volcengine_ecs_instance": tableVolcengineEcsInstance(ctx),
			"volcengine_ebs_volume":   tableVolcengineEbsVolume(ctx),
			// Networking
			"volcengine_vpc":                tableVolcengineVpc(ctx),
			"volcengine_vpc_subnet":         tableVolcengineVpcSubnet(ctx),
			"volcengine_vpc_security_group": tableVolcengineVpcSecurityGroup(ctx),
			"volcengine_eip":                tableVolcengineEip(ctx),
			// Load Balancers
			"volcengine_clb": tableVolcengineClb(ctx),
			"volcengine_alb": tableVolcengineAlb(ctx),
			// Gateway
			"volcengine_nat_gateway": tableVolcengineNatGateway(ctx),
			// Storage
			"volcengine_tos_bucket": tableVolcengineTosBucket(ctx),
			// Databases
			"volcengine_redis_instance":   tableVolcengineRedisInstance(ctx),
			"volcengine_mongodb_instance": tableVolcengineMongoDBInstance(ctx),
			// Kubernetes
			"volcengine_vke_cluster": tableVolcengineVkeCluster(ctx),
			// IAM
			"volcengine_iam_user":   tableVolcengineIamUser(ctx),
			"volcengine_iam_role":   tableVolcengineIamRole(ctx),
			"volcengine_iam_policy": tableVolcengineIamPolicy(ctx),
		},
	}
	return p
}
