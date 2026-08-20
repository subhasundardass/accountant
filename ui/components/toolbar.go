package components

import (
	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/theme/devtools"
)

func MainToolbar(p devtools.Painters) *toolbar.Widget {
	noop := func() {}

	return toolbar.New(
		toolbar.Items(
			toolbar.IconButton("Menu", icon.ExpMenu, noop),
			toolbar.Separator(),

			toolbar.IconButton("Dashboard", icon.ExpFolder, noop),
			toolbar.IconButton("Save", icon.ExpSave, noop),

			toolbar.Separator(),

			toolbar.IconButton("Search", icon.ExpSearch, noop),
			toolbar.IconButton("Settings", icon.ExpSettings, noop),
		),

		toolbar.Height(40),
		toolbar.ButtonSize(30),
		toolbar.Gap(10),
		toolbar.PainterOpt(p.Toolbar),
	)
}
