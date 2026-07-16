package volcengine

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/volcengine-go-sdk/service/alb"
	"github.com/volcengine/volcengine-go-sdk/service/clb"
	"github.com/volcengine/volcengine-go-sdk/service/ecs"
	"github.com/volcengine/volcengine-go-sdk/service/iam"
	"github.com/volcengine/volcengine-go-sdk/service/mongodb"
	"github.com/volcengine/volcengine-go-sdk/service/natgateway"
	"github.com/volcengine/volcengine-go-sdk/service/redis"
	"github.com/volcengine/volcengine-go-sdk/service/storageebs"
	"github.com/volcengine/volcengine-go-sdk/service/vke"
	"github.com/volcengine/volcengine-go-sdk/service/vpc"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// getSession creates or retrieves a cached volcengine session
func getSession(ctx context.Context, d *plugin.QueryData, region string) (*session.Session, error) {
	if region == "" {
		return nil, fmt.Errorf("region must be provided")
	}

	serviceCacheKey := fmt.Sprintf("session-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*session.Session), nil
	}

	ak, sk, err := getCredentials(ctx, d)
	if err != nil {
		return nil, err
	}

	config := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(ak, sk, "")).
		WithRegion(region)

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("error creating volcengine session for region %s: %v", region, err)
	}

	d.ConnectionManager.Cache.Set(serviceCacheKey, sess)
	return sess, nil
}

// ECSService returns the ECS service client for Volcengine
func ECSService(ctx context.Context, d *plugin.QueryData) (*ecs.ECS, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to ECSService")
	}

	serviceCacheKey := fmt.Sprintf("ecs-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*ecs.ECS), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := ecs.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// VPCService returns the VPC service client for Volcengine (includes EIP APIs)
func VPCService(ctx context.Context, d *plugin.QueryData) (*vpc.VPC, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to VPCService")
	}

	serviceCacheKey := fmt.Sprintf("vpc-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*vpc.VPC), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := vpc.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// CLBService returns the CLB service client for Volcengine
func CLBService(ctx context.Context, d *plugin.QueryData) (*clb.CLB, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to CLBService")
	}

	serviceCacheKey := fmt.Sprintf("clb-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*clb.CLB), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := clb.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// ALBService returns the ALB service client for Volcengine
func ALBService(ctx context.Context, d *plugin.QueryData) (*alb.ALB, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to ALBService")
	}

	serviceCacheKey := fmt.Sprintf("alb-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*alb.ALB), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := alb.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// MongoDBService returns the MongoDB service client for Volcengine
func MongoDBService(ctx context.Context, d *plugin.QueryData) (*mongodb.MONGODB, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to MongoDBService")
	}

	serviceCacheKey := fmt.Sprintf("mongodb-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*mongodb.MONGODB), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := mongodb.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// NatGatewayService returns the NAT Gateway service client for Volcengine
func NatGatewayService(ctx context.Context, d *plugin.QueryData) (*natgateway.NATGATEWAY, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to NatGatewayService")
	}

	serviceCacheKey := fmt.Sprintf("natgateway-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*natgateway.NATGATEWAY), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := natgateway.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// RedisService returns the Redis service client for Volcengine
func RedisService(ctx context.Context, d *plugin.QueryData) (*redis.REDIS, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to RedisService")
	}

	serviceCacheKey := fmt.Sprintf("redis-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*redis.REDIS), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := redis.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// VKEService returns the VKE (Kubernetes Engine) service client for Volcengine
func VKEService(ctx context.Context, d *plugin.QueryData) (*vke.VKE, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to VKEService")
	}

	serviceCacheKey := fmt.Sprintf("vke-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*vke.VKE), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := vke.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// StorageEBSService returns the EBS service client for Volcengine
func StorageEBSService(ctx context.Context, d *plugin.QueryData) (*storageebs.STORAGEEBS, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("region must be passed to StorageEBSService")
	}

	serviceCacheKey := fmt.Sprintf("storageebs-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*storageebs.STORAGEEBS), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := storageebs.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// IAMService returns the IAM service client for Volcengine (global service)
func IAMService(ctx context.Context, d *plugin.QueryData) (*iam.IAM, error) {
	region := GetDefaultRegion(d.Connection)

	serviceCacheKey := fmt.Sprintf("iam-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*iam.IAM), nil
	}

	sess, err := getSession(ctx, d, region)
	if err != nil {
		return nil, err
	}

	svc := iam.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

// TOSService returns the TOS (object storage) client for Volcengine
func TOSService(ctx context.Context, d *plugin.QueryData, region string) (*tos.ClientV2, error) {
	if region == "" {
		return nil, fmt.Errorf("region must be provided to TOSService")
	}

	serviceCacheKey := fmt.Sprintf("tos-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*tos.ClientV2), nil
	}

	ak, sk, err := getCredentials(ctx, d)
	if err != nil {
		return nil, err
	}

	// TOS endpoint format: tos-{region}.volces.com
	endpoint := fmt.Sprintf("tos-%s.volces.com", region)

	client, err := tos.NewClientV2(
		endpoint,
		tos.WithRegion(region),
		tos.WithCredentials(tos.NewStaticCredentials(ak, sk)),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating TOS client for region %s: %v", region, err)
	}

	d.ConnectionManager.Cache.Set(serviceCacheKey, client)
	return client, nil
}

// getCredentials returns the access key and secret key from config or environment
func getCredentials(_ context.Context, d *plugin.QueryData) (string, string, error) {
	config := GetConfig(d.Connection)

	var ak, sk string

	if config.AccessKey != nil {
		ak = *config.AccessKey
	} else {
		ak = os.Getenv("VOLCENGINE_ACCESS_KEY_ID")
		if ak == "" {
			ak = os.Getenv("VOLCENGINE_ACCESSKEY")
		}
	}

	if config.SecretKey != nil {
		sk = *config.SecretKey
	} else {
		sk = os.Getenv("VOLCENGINE_SECRET_ACCESS_KEY")
		if sk == "" {
			sk = os.Getenv("VOLCENGINE_SECRETKEY")
		}
	}

	if ak == "" || sk == "" {
		return "", "", fmt.Errorf("'access_key' and 'secret_key' must be set in the connection configuration or via VOLCENGINE_ACCESS_KEY_ID/VOLCENGINE_SECRET_ACCESS_KEY environment variables")
	}

	return ak, sk, nil
}

// GetDefaultRegion returns the default region used
func GetDefaultRegion(connection *plugin.Connection) string {
	config := GetConfig(connection)

	if config.Regions != nil && len(config.Regions) > 0 {
		region := config.Regions[0]
		if len(getInvalidRegions([]string{region})) > 0 {
			panic("\n\nConnection config has invalid region: " + region + ". Edit your connection configuration file and then restart Steampipe.")
		}
		return region
	}

	region := os.Getenv("VOLCENGINE_REGION")
	if region == "" {
		region = os.Getenv("VOLCENGINE_REGION_ID")
	}

	if region == "" {
		region = "cn-beijing"
	}

	return region
}

// getClientTimeout returns the configured timeout or 60s default
func getClientTimeout(d *plugin.QueryData) time.Duration {
	config := GetConfig(d.Connection)
	if config.Timeout != nil {
		return time.Duration(*config.Timeout) * time.Second
	}
	if envTimeout := os.Getenv("STEAMPIPE_VOLCENGINE_TIMEOUT"); envTimeout != "" {
		if seconds, err := strconv.Atoi(envTimeout); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 60 * time.Second
}
