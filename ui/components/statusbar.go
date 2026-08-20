package components

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/devtools"
)

func StatusBar(cs devtools.ColorScheme) *primitives.BoxWidget {
	const fontSize float32 = 12

	left := primitives.Text("Accountant").
		Color(cs.OnSurfaceSecondary).
		FontSize(fontSize)

	version := primitives.Text("v0.1.0").
		Color(cs.OnSurfaceSecondary).
		FontSize(fontSize)

	status := primitives.Text("Ready").
		Color(cs.Success).
		FontSize(fontSize)

	return primitives.HBox(
		left,
		status,
		primitives.Box().Width(0).Height(1),
		version,
	).
		Background(cs.Surface).
		PaddingXY(8, 4).
		Gap(16).
		Height(28).
		BorderStyle(1, cs.Border)
}
