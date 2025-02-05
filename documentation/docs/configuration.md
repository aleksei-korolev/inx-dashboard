---
description: This section describes the configuration parameters and their types for INX-Dashboard.
keywords:
- IOTA Node 
- Hornet Node
- Dashboard
- Configuration
- JSON
- Customize
- Config
- reference
---


# Core Configuration

INX-Dashboard uses a JSON standard format as a config file. If you are unsure about JSON syntax, you can find more information in the [official JSON specs](https://www.json.org).

You can change the path of the config file by using the `-c` or `--config` argument while executing `inx-dashboard` executable.

For example:
```bash
inx-dashboard -c config_defaults.json
```

You can always get the most up-to-date description of the config parameters by running:

```bash
inx-dashboard -h --full
```

## <a id="app"></a> 1. Application

| Name            | Description                                                                                            | Type    | Default value |
| --------------- | ------------------------------------------------------------------------------------------------------ | ------- | ------------- |
| checkForUpdates | Whether to check for updates of the application or not                                                 | boolean | true          |
| stopGracePeriod | The maximum time to wait for background processes to finish during shutdown before terminating the app | string  | "5m"          |

Example:

```json
  {
    "app": {
      "checkForUpdates": true,
      "stopGracePeriod": "5m"
    }
  }
```

## <a id="inx"></a> 2. INX

| Name    | Description                            | Type   | Default value    |
| ------- | -------------------------------------- | ------ | ---------------- |
| address | The INX address to which to connect to | string | "localhost:9029" |

Example:

```json
  {
    "inx": {
      "address": "localhost:9029"
    }
  }
```

## <a id="dashboard"></a> 3. Dashboard

| Name                      | Description                                                  | Type    | Default value           |
| ------------------------- | ------------------------------------------------------------ | ------- | ----------------------- |
| bindAddress               | The bind address on which the dashboard can be accessed from | string  | "localhost:8081"        |
| developerMode             | Whether to run the dashboard in dev mode                     | boolean | false                   |
| developerModeURL          | The URL to use for dev mode                                  | string  | "http://127.0.0.1:9090" |
| [auth](#dashboard_auth)   | Configuration for auth                                       | object  |                         |
| debugRequestLoggerEnabled | Whether the debug logging for requests should be enabled     | boolean | false                   |

### <a id="dashboard_auth"></a> Auth

| Name               | Description                                           | Type   | Default value                                                      |
| ------------------ | ----------------------------------------------------- | ------ | ------------------------------------------------------------------ |
| sessionTimeout     | How long the auth session should last before expiring | string | "72h"                                                              |
| username           | The auth username (max 25 chars)                      | string | "admin"                                                            |
| passwordHash       | The auth password+salt as a scrypt hash               | string | "0000000000000000000000000000000000000000000000000000000000000000" |
| passwordSalt       | The auth salt used for hashing the password           | string | "0000000000000000000000000000000000000000000000000000000000000000" |
| identityFilePath   | The path to the identity file used for JWT            | string | "identity.key"                                                     |
| identityPrivateKey | Private key used to sign the JWT tokens (optional)    | string | ""                                                                 |

Example:

```json
  {
    "dashboard": {
      "bindAddress": "localhost:8081",
      "developerMode": false,
      "developerModeURL": "http://127.0.0.1:9090",
      "auth": {
        "sessionTimeout": "72h",
        "username": "admin",
        "passwordHash": "0000000000000000000000000000000000000000000000000000000000000000",
        "passwordSalt": "0000000000000000000000000000000000000000000000000000000000000000",
        "identityFilePath": "identity.key",
        "identityPrivateKey": ""
      },
      "debugRequestLoggerEnabled": false
    }
  }
```

## <a id="profiling"></a> 4. Profiling

| Name        | Description                                       | Type    | Default value    |
| ----------- | ------------------------------------------------- | ------- | ---------------- |
| enabled     | Whether the profiling plugin is enabled           | boolean | false            |
| bindAddress | The bind address on which the profiler listens on | string  | "localhost:6060" |

Example:

```json
  {
    "profiling": {
      "enabled": false,
      "bindAddress": "localhost:6060"
    }
  }
```

## <a id="prometheus"></a> 5. Prometheus

| Name            | Description                                                     | Type    | Default value    |
| --------------- | --------------------------------------------------------------- | ------- | ---------------- |
| enabled         | Whether the prometheus plugin is enabled                        | boolean | false            |
| bindAddress     | The bind address on which the Prometheus HTTP server listens on | string  | "localhost:9312" |
| goMetrics       | Whether to include go metrics                                   | boolean | false            |
| processMetrics  | Whether to include process metrics                              | boolean | false            |
| promhttpMetrics | Whether to include promhttp metrics                             | boolean | false            |

Example:

```json
  {
    "prometheus": {
      "enabled": false,
      "bindAddress": "localhost:9312",
      "goMetrics": false,
      "processMetrics": false,
      "promhttpMetrics": false
    }
  }
```

