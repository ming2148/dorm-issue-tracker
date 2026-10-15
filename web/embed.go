package web

import "embed"

//go:embed index.html static/*
var WebFS embed.FS
