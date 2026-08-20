package ui

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/devtools"

	"github.com/subhasundardass/accountant/ui/components"
)

func Layout(
	cs devtools.ColorScheme,
	p devtools.Painters,
	workspace *primitives.BoxWidget,
) *primitives.BoxWidget {

	toolbar := components.MainToolbar(p)
	statusBar := components.StatusBar(cs)
	hotkeyPanel := components.HotkeyPanel(cs)

	separator := primitives.Box().
		Width(1).
		Background(cs.Border)

	mainArea := primitives.HBox(
		primitives.Expanded(workspace),
		separator,
		hotkeyPanel,
	)

	return primitives.VBox(
		toolbar,
		primitives.Expanded(mainArea),
		statusBar,
	)
}
