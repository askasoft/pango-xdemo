package server

import (
	"github.com/askasoft/pango-xdemo/app/tasks"
	"github.com/askasoft/pango/cog"
)

var schedules = cog.NewLinkedHashMap[string, func()](
	cog.KV("cleanUploadFiles", tasks.CleanUploadFiles),
)
