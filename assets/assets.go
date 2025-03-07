// Package assets contains the assets that are embedded in the storctl tool.
// It includes the playbooks and templates for the labs.
// It also includes the example config files for the labs.
package assets

import (
	"embed"
)

//go:embed playbooks/*
var PlaybookFiles embed.FS

//go:embed templates/*
var TemplateFiles embed.FS

//go:embed virt/templates/*
var VirtTemplateFiles embed.FS
