package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/iam"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableVolcengineIamPolicy(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_iam_policy",
		Description: "Volcengine IAM Policy",
		List: &plugin.ListConfig{
			Hydrate: listIamPolicies,
			Tags:    map[string]string{"service": "iam", "action": "ListPolicies"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("policy_name"),
			Hydrate:    getIamPolicy,
			Tags:       map[string]string{"service": "iam", "action": "GetPolicy"},
		},
		Columns: []*plugin.Column{
			{
				Name:        "policy_name",
				Type:        proto.ColumnType_STRING,
				Description: "The name of the policy.",
			},
			{
				Name:        "policy_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the policy.",
			},
			{
				Name:        "arn",
				Type:        proto.ColumnType_STRING,
				Description: "The ARN of the policy.",
			},
			{
				Name:        "policy_type",
				Type:        proto.ColumnType_STRING,
				Description: "The type of the policy (custom/system).",
			},
			{
				Name:        "description",
				Type:        proto.ColumnType_STRING,
				Description: "The description of the policy.",
			},
			{
				Name:        "default_version",
				Type:        proto.ColumnType_STRING,
				Description: "The default version of the policy.",
			},
			{
				Name:        "attachment_count",
				Type:        proto.ColumnType_INT,
				Description: "Number of entities the policy is attached to.",
			},
			{
				Name:        "create_date",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the policy was created.",
			},
			{
				Name:        "update_date",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the policy was last updated.",
			},
			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PolicyName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Arn").Transform(ensureStringArray),
			},
		},
	}
}

func listIamPolicies(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := IAMService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_policy.listIamPolicies", "connection_error", err)
		return nil, err
	}

	input := &iam.ListPoliciesInput{}

	if value, ok := GetStringQualValue(d.Quals, "policy_name"); ok && value != nil {
		input.PolicyName = value
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.ListPolicies(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_iam_policy.listIamPolicies", "query_error", err)
			return nil, err
		}

		if response.Policies == nil {
			break
		}

		for _, policy := range response.Policies {
			d.StreamListItem(ctx, policy)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if response.Marker == nil || *response.Marker == "" {
			break
		}
		input.Marker = response.Marker
	}

	return nil, nil
}

func getIamPolicy(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := IAMService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_policy.getIamPolicy", "connection_error", err)
		return nil, err
	}

	var policyName string
	if h.Item != nil {
		policy := h.Item.(*iam.Policy)
		policyName = *policy.PolicyName
	} else {
		policyName = d.EqualsQuals["policy_name"].GetStringValue()
	}

	if policyName == "" {
		return nil, nil
	}

	input := &iam.GetPolicyInput{
		PolicyName: volcengine.String(policyName),
	}

	response, err := client.GetPolicy(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_policy.getIamPolicy", "query_error", err)
		return nil, err
	}

	return response.Policy, nil
}
