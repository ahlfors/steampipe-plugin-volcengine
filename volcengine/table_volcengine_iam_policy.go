package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/iam"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
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
				Transform:   transform.FromField("PolicyName"),
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
				Transform:   transform.FromField("PolicyTrn"),
			},
			{
				Name:        "policy_type",
				Type:        proto.ColumnType_STRING,
				Description: "The type of the policy (custom/system).",
				Transform:   transform.FromField("PolicyType"),
			},
			{
				Name:        "description",
				Type:        proto.ColumnType_STRING,
				Description: "The description of the policy.",
				Transform:   transform.FromField("Description"),
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
				Type:        proto.ColumnType_STRING,
				Description: "The time when the policy was created.",
				Transform:   transform.FromField("CreateDate"),
			},
			{
				Name:        "update_date",
				Type:        proto.ColumnType_STRING,
				Description: "The time when the policy was last updated.",
				Transform:   transform.FromField("UpdateDate"),
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
				Transform:   transform.FromField("PolicyTrn").Transform(ensureStringArray),
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
	limit := int32(100)
	input.Limit = &limit

	if value, ok := GetStringQualValue(d.Quals, "policy_type"); ok && value != nil {
		input.Scope = value
	}

	offset := int32(0)
	for {
		d.WaitForListRateLimit(ctx)
		input.Offset = &offset
		response, err := client.ListPolicies(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_iam_policy.listIamPolicies", "query_error", err)
			return nil, err
		}

		if response.PolicyMetadata == nil {
			break
		}

		for _, policy := range response.PolicyMetadata {
			d.StreamListItem(ctx, policy)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(response.PolicyMetadata) == 0 {
			break
		}

		offset += int32(len(response.PolicyMetadata))
		if response.Total != nil && offset >= *response.Total {
			break
		}
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
	var policyType string
	if h.Item != nil {
		policy := h.Item.(*iam.PolicyMetadatumForListPoliciesOutput)
		if policy.PolicyName != nil {
			policyName = *policy.PolicyName
		}
		if policy.PolicyType != nil {
			policyType = *policy.PolicyType
		}
	} else {
		policyName = d.EqualsQuals["policy_name"].GetStringValue()
		policyType = d.EqualsQuals["policy_type"].GetStringValue()
		if policyType == "" {
			policyType = "Custom"
		}
	}

	if policyName == "" || policyType == "" {
		return nil, nil
	}

	input := &iam.GetPolicyInput{
		PolicyName: volcengine.String(policyName),
		PolicyType: volcengine.String(policyType),
	}

	response, err := client.GetPolicy(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_policy.getIamPolicy", "query_error", err)
		return nil, err
	}

	return response.Policy, nil
}
