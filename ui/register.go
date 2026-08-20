package ui

import (
	"fmt"

	"github.com/gogpu/ui/widget"
	"github.com/subhasundardass/accountant/internal/context"
)

type Screen struct {
	ID     string
	Title  string
	Render func(ctx *context.AppContext) widget.Widget
}

var routes = map[string]Screen{}

func Register(key string, screen Screen) {
	if _, exists := routes[key]; exists {
		panic("screen already registered: " + key)
	}

	if screen.ID != key {
		panic(fmt.Sprintf(
			"screen ID %q does not match registration key %q",
			screen.ID,
			key,
		))
	}

	routes[key] = screen
}

func GetScreen(id string) (Screen, bool) {
	screen, ok := routes[id]
	return screen, ok
}

// OpenScreen renders a registered screen and puts it
// into the application workspace.
func OpenScreen(id string, ctx *context.AppContext) bool {
	screen, ok := GetScreen(id)
	if !ok {
		return false
	}

	content := screen.Render(ctx)
	if content == nil {
		return false
	}

	SetWorkspaceContent(content)

	return true
}
