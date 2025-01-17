package assets

import (
	"embed"
)

//go:embed playbooks/*
var PlaybookFiles embed.FS

//go:embed templates/*
var TemplateFiles embed.FS
