[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-consul-check)
![Go Test](https://github.com/sensu/sensu-consul-check/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/sensu/sensu-consul-check/workflows/goreleaser/badge.svg)

# Check Consul service health

## Table of Contents
- [Overview](#overview)
- [Usage examples](#usage-examples)
  - [Help output](#help-output)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Contributing](#contributing)

## Overview

The sensu-consul-check is a [Sensu Check][2] that queries the health
of Consul service health checks.  It is a golang update of the [check-consul-service-health.rb][5]
script in the [sensu-plugins-consul][6] repository.

## Usage examples

### Help output
```
Usage:
  sensu-consul-check [flags]
  sensu-consul-check [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -A, --acl-token string          ACL token for connecting to Consul
  -a, --all                       Get all services (not compatible with --tags)
  -c, --consul-server string      Consul server URL (default "http://127.0.0.1:8500")
  -x, --exclude-service strings   Service managed by Consul to exclude from check
  -f, --fail-if-not-found         Fail if no service is found
  -h, --help                      help for sensu-consul-check
  -i, --insecure-skip-verify      Skip TLS certificate verification (not recommended!)
  -n, --node string               Check all Consul service running on the specified node
  -s, --service string            Service managed by Consul to check (default "consul")
  -t, --tags strings              Filter services by a comma-separated list of tags (requires --service)
  -T, --trusted-ca-file string    TLS CA certificate bundle in PEM format

Use "sensu-consul-check [command] --help" for more information about a command.
```

## Configuration

### Asset registration

[Sensu Assets][3] are the best way to make use of this plugin. If you're not
using an asset, please consider doing so! If you're using sensuctl 5.13 with
Sensu Backend 5.13 or later, you can use the following command to add the asset:

```
sensuctl asset add sensu/sensu-consul-check
```

If you're using an earlier version of sensuctl, you can find the asset on the
[Bonsai Asset Index][4].

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: sensu-consul-check
  namespace: default
spec:
  command: >-
    sensu-consul-check
    --consul-server http://127.0.0.1:8500
    --all
  subscriptions:
  - system
  runtime_assets:
  - sensu/sensu-consul-check
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an
Asset. If you would like to compile and install the plugin from source or
contribute to it, download the latest version or create an executable from
this source.

From the local path of the sensu-consul-check repository:

```
go build
```

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[3]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[4]: https://bonsai.sensu.io/assets/sensu/sensu-consul-check
[5]: https://github.com/sensu-plugins/sensu-plugins-consul/blob/master/bin/sensu-consul-check.rb
[6]: https://github.com/sensu-plugins/sensu-plugins-consul
