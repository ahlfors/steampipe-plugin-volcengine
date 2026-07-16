package main

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/ahlfors/steampipe-plugin-volcengine/volcengine"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{PluginFunc: volcengine.Plugin})
}
