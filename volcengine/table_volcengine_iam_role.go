package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/iam"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableVolcengineIamRole(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_iam_role",
		Description: "Volcengine IAM Role",
		List: &plugin.ListConfig{
			Hydrate: listIamRoles,
			Tags:    map[string]string{"service": "iam", "action": "ListRoles"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("role_name"),
			Hydrate:    getIamRole,
			Tags:       map[string]string{"service": "iam", "action": "GetRole"},
		},
		Columns: []*plugin.Column{
			{
				Name:        "role_name",
				Type:        proto.ColumnType_STRING,
				Description: "The name of the IAM role.",
			},
			{
				Name:        "role_id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the IAM role.",
			},
			{
				Name:        "arn",
				Type:        proto.ColumnType_STRING,
				Description: "The ARN of the IAM role.",
			},
			{
				Name:        "display_name",
				Type:        proto.ColumnType_STRING,
				Description: "The display name of the role.",
			},
			{
				Name:        "description",
				Type:        proto.ColumnType_STRING,
				Description: "The description of the role.",
			},
			{
				Name:        "trust_policy_document",
				Type:        proto.ColumnType_STRING,
				Description: "The trust policy document attached to the role.",
			},
			{
				Name:        "create_date",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the role was created.",
			},
			{
				Name:        "update_date",
				Type:        proto.ColumnType_TIMESTAMP,
				Description: "The time when the role was last updated.",
			},
			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RoleName"),
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

func listIamRoles(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := IAMService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_role.listIamRoles", "connection_error", err)
		return nil, err
	}

	input := &iam.ListRolesInput{}

	if value, ok := GetStringQualValue(d.Quals, "role_name"); ok && value != nil {
		input.RoleName = value
	}

	for {
		d.WaitForListRateLimit(ctx)
		response, err := client.ListRoles(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_iam_role.listIamRoles", "query_error", err)
			return nil, err
		}

		if response.Roles == nil {
			break
		}

		for _, role := range response.Roles {
			d.StreamListItem(ctx, role)
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

func getIamRole(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := IAMService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_role.getIamRole", "connection_error", err)
		return nil, err
	}

	var roleName string
	if h.Item != nil {
		role := h.Item.(*iam.Role)
		roleName = *role.RoleName
	} else {
		roleName = d.EqualsQuals["role_name"].GetStringValue()
	}

	if roleName == "" {
		return nil, nil
	}

	input := &iam.GetRoleInput{
		RoleName: volcengine.String(roleName),
	}

	response, err := client.GetRole(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_role.getIamRole", "query_error", err)
		return nil, err
	}

	return response.Role, nil
}
