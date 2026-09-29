package volcengine

import "github.com/turbot/steampipe-plugin-sdk/v6/plugin"

// volcengineConfig holds the connection configuration for Volcengine.
type volcengineConfig struct {
	Regions          []string `hcl:"regions,optional"`
	AccessKey        *string  `hcl:"access_key"`
	SecretKey        *string  `hcl:"secret_key"`
	IgnoreErrorCodes []string `hcl:"ignore_error_codes,optional"`
	Timeout          *int     `hcl:"timeout,optional"`
}

// ConfigInstance returns a new instance of the connection config.
func ConfigInstance() interface{} {
	return &volcengineConfig{}
}

// GetConfig retrieves and casts connection config from query data.
func GetConfig(connection *plugin.Connection) volcengineConfig {
	if connection == nil {
		return volcengineConfig{}
	}
	config, _ := connection.GetConfig().(volcengineConfig)
	return config
}
