package testtemplates

import "embed"

// content holds our static web server content.
//
//go:embed *.tpl.html
var Content embed.FS
