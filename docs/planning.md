# Project planning

## Commands

* `get` - get the list of lab environments currently running, or a specific lab environment by its name (think `kubectl get`)
* `create` - create a new lab environment
* `delete` - destroy an existing lab environment
* `describe` - show detailed information about a specific lab environment (node specs, IP addresses, installed components, etc.)
* `logs` - fetch logs from the deployment process (terraform, ansible) for troubleshooting
* `status` - show the current state of lab environment components (k8s cluster health, MinIO status, etc.)
* `config` - manage configuration for the CLI tool (cloud credentials, default settings, etc.)
* `apply` - apply changes to an existing lab environment using a configuration file
* `scale` - adjust the size of an existing lab environment (add/remove nodes)
* `exec` - execute commands directly on the lab environment (similar to kubectl exec)
* `port-forward` - create port forwarding to access services running in the lab
* `backup` - create backups of lab environment state and data
* `restore` - restore a lab environment from a backup
* `list-templates` - show available templates/configurations for lab environments
* `validate` - validate a lab environment configuration file before applying it

