package volcengine

import (
	"context"

	"github.com/volcengine/volcengine-go-sdk/service/iam"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

func tableVolcengineIamUser(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "volcengine_iam_user",
		Description: "Volcengine IAM User",
		List: &plugin.ListConfig{
			Hydrate: listIamUsers,
			Tags:    map[string]string{"service": "iam", "action": "ListUsers"},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("user_name"),
			Hydrate:    getIamUser,
			Tags:       map[string]string{"service": "iam", "action": "GetUser"},
		},
		Columns: []*plugin.Column{
			{
				Name:        "user_name",
				Type:        proto.ColumnType_STRING,
				Description: "The name of the IAM user.",
				Transform:   transform.FromField("UserName"),
			},
			{
				Name:        "user_id",
				Type:        proto.ColumnType_INT,
				Description: "The ID of the IAM user.",
				Transform:   transform.FromField("Id"),
			},
			{
				Name:        "arn",
				Type:        proto.ColumnType_STRING,
				Description: "The ARN of the IAM user.",
				Transform:   transform.FromField("Trn"),
			},
			{
				Name:        "display_name",
				Type:        proto.ColumnType_STRING,
				Description: "The display name of the user.",
				Transform:   transform.FromField("DisplayName"),
			},
			{
				Name:        "email",
				Type:        proto.ColumnType_STRING,
				Description: "The email address of the user.",
				Transform:   transform.FromField("Email"),
			},
			{
				Name:        "mobile_phone",
				Type:        proto.ColumnType_STRING,
				Description: "The mobile phone number of the user.",
				Transform:   transform.FromField("MobilePhone"),
			},
			{
				Name:        "description",
				Type:        proto.ColumnType_STRING,
				Description: "The description of the user.",
				Transform:   transform.FromField("Description"),
			},
			{
				Name:        "create_date",
				Type:        proto.ColumnType_STRING,
				Description: "The time when the user was created.",
				Transform:   transform.FromField("CreateDate"),
			},
			{
				Name:        "update_date",
				Type:        proto.ColumnType_STRING,
				Description: "The time when the user was last updated.",
				Transform:   transform.FromField("UpdateDate"),
			},
			{
				Name:        "account_id",
				Type:        proto.ColumnType_STRING,
				Description: "The account ID to which the user belongs.",
				Transform:   transform.FromField("AccountId").Transform(int64PtrToString),
			},
			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserName"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Trn").Transform(ensureStringArray),
			},
		},
	}
}

func listIamUsers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := IAMService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_user.listIamUsers", "connection_error", err)
		return nil, err
	}

	input := &iam.ListUsersInput{}
	limit := int32(100)
	input.Limit = &limit

	if value, ok := GetStringQualValue(d.Quals, "user_name"); ok && value != nil {
		input.Query = value
	}

	offset := int32(0)
	for {
		d.WaitForListRateLimit(ctx)
		input.Offset = &offset
		response, err := client.ListUsers(input)
		if err != nil {
			plugin.Logger(ctx).Error("volcengine_iam_user.listIamUsers", "query_error", err)
			return nil, err
		}

		if response.UserMetadata == nil {
			break
		}

		for _, user := range response.UserMetadata {
			d.StreamListItem(ctx, user)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(response.UserMetadata) == 0 {
			break
		}

		offset += int32(len(response.UserMetadata))
		if response.Total != nil && offset >= *response.Total {
			break
		}
	}

	return nil, nil
}

func getIamUser(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	client, err := IAMService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_user.getIamUser", "connection_error", err)
		return nil, err
	}

	var userName string
	if h.Item != nil {
		user := h.Item.(*iam.UserMetadataForListUsersOutput)
		if user.UserName != nil {
			userName = *user.UserName
		}
	} else {
		userName = d.EqualsQuals["user_name"].GetStringValue()
	}

	if userName == "" {
		return nil, nil
	}

	input := &iam.GetUserInput{
		UserName: volcengine.String(userName),
	}

	response, err := client.GetUser(input)
	if err != nil {
		plugin.Logger(ctx).Error("volcengine_iam_user.getIamUser", "query_error", err)
		return nil, err
	}

	return response.User, nil
}
