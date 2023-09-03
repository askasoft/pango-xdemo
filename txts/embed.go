package txts

import (
	"embed"
)

// FS embed text message folder
//
//go:embed *.ini */*
var FS embed.FS
