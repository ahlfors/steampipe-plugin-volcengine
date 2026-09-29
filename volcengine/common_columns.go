package volcengine

import (
	"context"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v6/memoize"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/volcengine/volcengine-go-sdk/service/iam"
)

// volcengineCommonColumnData stores the common column data
type volcengineCommonColumnData struct {
	AccountID string
}

// getAccountId returns the account ID for the connection
func getAccountId(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	data, err := getAccountDetails(ctx, d, h)
	if err != nil {
		return nil, err
	}
	return data.(*volcengineCommonColumnData).AccountID, nil
}

// getCommonColumns returns common column data (account_id)
func getCommonColumns(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	data, err := getAccountDetails(ctx, d, h)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var getAccountDetailsMemoize = plugin.HydrateFunc(getAccountDetailsUncached).Memoize(memoize.WithCacheKeyFunction(getAccountDetailsCacheKey))

func getAccountDetailsCacheKey(_ context.Context, _ *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	return "volcengine-account-details", nil
}

func getAccountDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	config, err := getAccountDetailsMemoize(ctx, d, h)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// getAccountDetailsUncached retrieves the account ID by calling Volcengine IAM APIs.
// Strategy:
//  1. Call IAM ListAccessKeys (without UserName) to discover the current caller's UserName.
//  2. Call IAM GetUser with that UserName to retrieve the AccountId field.
//
// If either call fails the function falls back to an empty string so that
// queries are not blocked when IAM permissions are restricted.
func getAccountDetailsUncached(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	commonData := &volcengineCommonColumnData{}

	iamSvc, err := IAMService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Warn("getAccountDetailsUncached", "iam_service_error", err)
		return commonData, nil
	}

	// Step 1: ListAccessKeys (current identity) → UserName
	listAKOutput, err := iamSvc.ListAccessKeys(nil)
	if err != nil {
		plugin.Logger(ctx).Warn("getAccountDetailsUncached", "list_access_keys_error", err)
		return commonData, nil
	}

	if len(listAKOutput.AccessKeyMetadata) == 0 {
		plugin.Logger(ctx).Warn("getAccountDetailsUncached", "msg", "no access keys returned, cannot determine account_id")
		return commonData, nil
	}

	userName := listAKOutput.AccessKeyMetadata[0].UserName
	if userName == nil || *userName == "" {
		plugin.Logger(ctx).Warn("getAccountDetailsUncached", "msg", "UserName not available in ListAccessKeys response")
		return commonData, nil
	}

	// Step 2: GetUser → AccountId
	getUserOutput, err := iamSvc.GetUser(&iam.GetUserInput{UserName: userName})
	if err != nil {
		plugin.Logger(ctx).Warn("getAccountDetailsUncached", "get_user_error", err)
		return commonData, nil
	}

	if getUserOutput.User != nil && getUserOutput.User.AccountId != nil {
		commonData.AccountID = strconv.FormatInt(*getUserOutput.User.AccountId, 10)
	}

	return commonData, nil
}
