package main

import (
	"fmt"
	"net/url"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	ConsulServer       string
	Node               string
	Service            string
	Tags               []string
	All                bool
	Token              string
	FailIfNotFound     bool
	InsecureSkipVerify bool
	TrustedCAFile      string
	ExcludeService     []string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-consul-check",
			Short:    "Consul Service Health Check",
			Keyspace: "sensu.io/plugins/sensu-consul-check/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "consul-server",
			Env:       "",
			Argument:  "consul-server",
			Shorthand: "c",
			Default:   "http://127.0.0.1:8500",
			Usage:     "Consul server URL",
			Value:     &plugin.ConsulServer,
		},
		{
			Path:      "node",
			Env:       "",
			Argument:  "node",
			Shorthand: "n",
			Default:   "",
			Usage:     "Check all Consul service running on the specified node",
			Value:     &plugin.Node,
		},
		{
			Path:      "service",
			Env:       "",
			Argument:  "service",
			Shorthand: "s",
			Default:   "consul",
			Usage:     "Service managed by Consul to check",
			Value:     &plugin.Service,
		},
		{
			Path:      "tags",
			Env:       "",
			Argument:  "tags",
			Shorthand: "t",
			Default:   []string{},
			Usage:     "Filter services by a comma-separated list of tags (requires --service)",
			Value:     &plugin.Tags,
		},
		{
			Path:      "all",
			Env:       "",
			Argument:  "all",
			Shorthand: "a",
			Default:   false,
			Usage:     "Get all services (not compatible with --tags)",
			Value:     &plugin.All,
		},
		{
			Path:      "fail-if-not-found",
			Env:       "",
			Argument:  "fail-if-not-found",
			Shorthand: "f",
			Default:   false,
			Usage:     "Fail if no service is found",
			Value:     &plugin.FailIfNotFound,
		},
		{
			Path:      "insecure-skip-verify",
			Env:       "",
			Argument:  "insecure-skip-verify",
			Shorthand: "i",
			Default:   false,
			Usage:     "Skip TLS certificate verification (not recommended!)",
			Value:     &plugin.InsecureSkipVerify,
		},
		{
			Path:      "trusted-ca-file",
			Env:       "",
			Argument:  "trusted-ca-file",
			Shorthand: "T",
			Default:   "",
			Usage:     "TLS CA certificate bundle in PEM format",
			Value:     &plugin.TrustedCAFile,
		},
		{
			Path:      "acl-token",
			Env:       "",
			Argument:  "acl-token",
			Shorthand: "A",
			Default:   "",
			Usage:     "ACL token for connecting to Consul",
			Value:     &plugin.Token,
		},
		{
			Path:      "exclude-service",
			Env:       "",
			Argument:  "exclude-service",
			Shorthand: "x",
			Default:   []string{},
			Usage:     "Service managed by Consul to exclude from check",
			Value:     &plugin.ExcludeService,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if len(plugin.Tags) > 0 && plugin.All {
		return sensu.CheckStateCritical, fmt.Errorf("--tags and --all are mutually exclusive")
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	conf := consul.DefaultConfig()
	if strings.HasPrefix(plugin.ConsulServer, "http://") || strings.HasPrefix(plugin.ConsulServer, "https://") {
		url, err := url.Parse(plugin.ConsulServer)
		if err != nil {
			return sensu.CheckStateCritical, fmt.Errorf("Failed to parse consul server URL %s: %v", plugin.ConsulServer, err)
		}
		conf.Address = url.Host
		conf.Scheme = url.Scheme
	} else {
		conf.Address = plugin.ConsulServer
	}

	if conf.Scheme == "https" {
		conf.TLSConfig.InsecureSkipVerify = plugin.InsecureSkipVerify
		if len(plugin.TrustedCAFile) > 0 {
			conf.TLSConfig.CAFile = plugin.TrustedCAFile
		}
	}

	if len(plugin.Token) > 0 {
		conf.Token = plugin.Token
	}

	client, err := consul.NewClient(conf)
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("Failed to create Consul client: %v", err)
	}
	health := client.Health()

	var healthChecks consul.HealthChecks

	if len(plugin.Service) > 0 && len(plugin.Tags) > 0 {
		// Future support for filters?
		// serviceEntries, _, err := health.Service(plugin.Service, plugin.Tags, false, &QueryOptions{Filter: "Node == foo and tag1 in ServiceTags"})
		serviceEntries, _, err := health.ServiceMultipleTags(plugin.Service, plugin.Tags, false, nil)
		if err != nil {
			return sensu.CheckStateCritical, fmt.Errorf("Failed to get Service health for %q: %v", plugin.Service, err)
		}
		for _, v := range serviceEntries {
			healthChecks = append(healthChecks, v.Checks...)
		}
	} else if len(plugin.Node) > 0 {
		var err error
		healthChecks, _, err = health.Node(plugin.Node, nil)
		if err != nil {
			return sensu.CheckStateCritical, fmt.Errorf("Failed to get health checks for node %q: %v", plugin.Node, err)
		}
	} else if plugin.All {
		var err error
		healthChecks, _, err = health.State("any", nil)
		if err != nil {
			return sensu.CheckStateCritical, fmt.Errorf("Failed to get health checks for \"any\": %v", err)
		}
	} else if len(plugin.Service) > 0 {
		serviceEntries, _, err := health.Service(plugin.Service, "", false, nil)
		if err != nil {
			return sensu.CheckStateCritical, fmt.Errorf("Failed to get Service health for %q: %v", plugin.Service, err)
		}
		for _, v := range serviceEntries {
			healthChecks = append(healthChecks, v.Checks...)
		}
	}

	var (
		found     bool
		warnings  int
		criticals int
		skip      bool
	)

	for _, v := range healthChecks {
		found = true
		skip = false

		if len(plugin.ExcludeService) > 0 {
			for _, s := range plugin.ExcludeService {
				if v.ServiceName == s {
					skip = true
				}
			}
		}

		if v.Status == "passing" {
			continue
		}

		if skip {
			continue
		}

		switch v.Status {
		case "critical", "unknown":
			criticals++
			fmt.Printf("%s CRITICAL: %s on %s\n", plugin.PluginConfig.Name, v.CheckID, v.Node)
		case "warning":
			warnings++
			fmt.Printf("%s WARNING: %s on %s\n", plugin.PluginConfig.Name, v.CheckID, v.Node)
		}
	}

	if !found && plugin.FailIfNotFound {
		fmt.Printf("%s CRITICAL: no checks found for provided arguments\n", plugin.PluginConfig.Name)
		return sensu.CheckStateCritical, nil
	}
	if criticals > 0 {
		return sensu.CheckStateCritical, nil
	} else if warnings > 0 {
		return sensu.CheckStateWarning, nil
	}

	if found {
		fmt.Printf("%s OK: All Consul service checks are passing\n", plugin.PluginConfig.Name)
	} else {
		fmt.Printf("%s OK: no checks found for provided arguments\n", plugin.PluginConfig.Name)
	}
	return sensu.CheckStateOK, nil
}
