package components

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/devtools"
)

func HotkeyPanel(cs devtools.ColorScheme) *primitives.BoxWidget {
	buttons := primitives.VBox(
		hotkeyButton(cs, "F1", "Dashboard"),
		hotkeyButton(cs, "F2", "Accounts"),
		hotkeyButton(cs, "F3", "Sales"),
		hotkeyButton(cs, "F4", "Purchase"),
		hotkeyButton(cs, "F5", "Inventory"),
		hotkeyButton(cs, "F6", "Ledger"),
		hotkeyButton(cs, "F7", "Reports"),

		primitives.Box().
			Height(12),

		hotkeyButton(cs, "F10", "Search"),
		hotkeyButton(cs, "F12", "Settings"),
	).
		Gap(4).
		Padding(8)

	return primitives.Box(
		buttons,
	).
		Width(200).
		Background(cs.Surface)
}

func hotkeyButton(
	cs devtools.ColorScheme,
	key string,
	label string,
) *primitives.BoxWidget {
	keyText := primitives.Text(key).
		Color(cs.Primary).
		FontSize(12)

	labelText := primitives.Text(label).
		Color(cs.OnSurface).
		FontSize(13)

	return primitives.HBox(
		keyText,
		labelText,
	).
		Height(36).
		Gap(12).
		PaddingXY(8, 4)
}
