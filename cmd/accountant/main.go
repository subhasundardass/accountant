package main

import (
	"log"

	_ "github.com/gogpu/gg/gpu"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/devtools"
	coreapp "github.com/subhasundardass/accountant/internal/app"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/ui"
)

func main() {

	// -------------------------------------------------------------------------
	// Bootstrap
	// -------------------------------------------------------------------------

	b, err := coreapp.NewBootstrap()
	if err != nil {
		return
	}

	// Always clean up application resources when the UI exits.
	defer func() {
		if err := b.Shutdown(); err != nil {
			logger.Errorf(
				"Application shutdown error: %v",
				err,
			)
		}
	}()

	// -------------------------------------------------------------------------
	// Theme
	// -------------------------------------------------------------------------

	dt := devtools.NewTheme()
	cs := dt.Colors
	p := devtools.NewPainters(dt)

	// -------------------------------------------------------------------------
	// Application workspace
	// -------------------------------------------------------------------------

	workspace := buildWorkspace(
		// b,
		cs,
		// p,
	)

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("My App").
		WithSize(800, 600))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(dt.AsTheme()),
	)

	root := ui.Layout(
		cs,
		p,
		workspace,
	)

	uiApp.SetRoot(
		root,
		// Background(widget.RGBA8(255, 255, 255, 255)),
	)

	// -- Global Hotkey Event
	gogpuApp.EventSource().OnKeyPress(
		func(key gpucontext.Key, mods gpucontext.Modifiers) {
			b.Hotkeys.Handle(key, mods)
		},
	)

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

// -------------------------------------------------------------------------
// Workspace
// -------------------------------------------------------------------------
func buildWorkspace(cs devtools.ColorScheme) *primitives.BoxWidget {
	return primitives.Box(
		ui.NewWorkspaceSlot(),
	).Background(cs.Background)
}
