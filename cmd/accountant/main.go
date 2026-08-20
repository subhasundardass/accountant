// package main

// import (
// 	"log"

// 	"github.com/gogpu/gogpu"
// 	// "github.com/gogpu/ui/app"
// 	"github.com/gogpu/ui/app"
// 	"github.com/gogpu/ui/desktop"
// 	"github.com/gogpu/ui/primitives"
// 	"github.com/gogpu/ui/widget"
// 	// coreapp "github.com/subhasundardass/accountant/internal/app"
// )

// func main() {

// 	// -------------------------------------------------------------------------
// 	// Bootstrap
// 	// -------------------------------------------------------------------------

// 	// b, err := coreapp.NewBootstrap()
// 	// if err != nil {
// 	// 	return
// 	// }

// 	// // Always clean up application resources when the UI exits.
// 	// defer func() {
// 	// 	if err := b.Shutdown(); err != nil {
// 	// 		logger.Errorf(
// 	// 			"Application shutdown error: %v",
// 	// 			err,
// 	// 		)
// 	// 	}
// 	// }()

// 	// -------------------------------------------------------------------------
// 	// Theme
// 	// -------------------------------------------------------------------------

// 	// dt := devtools.NewTheme()
// 	// cs := dt.Colors
// 	// p := devtools.NewPainters(dt)

// 	// -------------------------------------------------------------------------
// 	// Application workspace
// 	// -------------------------------------------------------------------------

// 	// workspace := buildWorkspace(
// 	// 	b,
// 	// 	cs,
// 	// 	p,
// 	// )

// 	// -------------------------------------------------------------------------
// 	// Main window
// 	// -------------------------------------------------------------------------

// 	// gogpuApp := gogpu.NewApp(
// 	// 	gogpu.DefaultConfig().
// 	// 		WithTitle("Accountant").
// 	// 		WithSize(800, 600).
// 	// 		WithResizable(true),
// 	// )
// 	// -------------------------------------------------------------------------
// 	// gogpu/ui application
// 	// -------------------------------------------------------------------------

// 	// uiApp := app.New(
// 	// 	app.WithWindowProvider(gogpuApp),
// 	// 	app.WithPlatformProvider(gogpuApp),
// 	// 	app.WithEventSource(gogpuApp.EventSource()),
// 	// )

// 	// -------------------------------------------------------------------------
// 	// Root layout
// 	// -------------------------------------------------------------------------

// 	// root := ui.Layout(
// 	// 	cs,
// 	// 	p,
// 	// 	workspace,
// 	// )

// 	// uiApp.SetRoot(root)
// 	// uiApp.SetRoot(
// 	// 	primitives.Box(
// 	// 		primitives.Text("Hello gogpu/ui!").
// 	// 			FontSize(24).Bold().
// 	// 			Color(widget.RGBA8(33, 33, 33, 255)),
// 	// 		primitives.Text("Enterprise-grade GUI for Go").
// 	// 			FontSize(16).
// 	// 			Color(widget.RGBA8(100, 100, 100, 255)),
// 	// 	).Padding(24).Gap(12).Background(widget.RGBA8(255, 255, 255, 255)),
// 	// )

// 	// if err := desktop.Run(gogpuApp, uiApp); err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
// 		WithTitle("My App").
// 		WithSize(800, 600))

// 	uiApp := app.New(
// 		app.WithWindowProvider(gogpuApp),
// 		app.WithPlatformProvider(gogpuApp),
// 		app.WithEventSource(gogpuApp.EventSource()),
// 	)

// 	uiApp.SetRoot(
// 		primitives.Box(
// 			primitives.Text("Hello gogpu/ui!").
// 				FontSize(24).Bold().
// 				Color(widget.RGBA8(33, 33, 33, 255)),
// 			primitives.Text("Enterprise-grade GUI for Go").
// 				FontSize(16).
// 				Color(widget.RGBA8(100, 100, 100, 255)),
// 		).Padding(24).Gap(12).Background(widget.RGBA8(255, 255, 255, 255)),
// 	)

// 	if err := desktop.Run(gogpuApp, uiApp); err != nil {
// 		log.Fatal(err)
// 	}
// }

// // -------------------------------------------------------------------------
// // Workspace
// // -------------------------------------------------------------------------
// // func buildWorkspace(
// // 	b *coreapp.Bootstrap,
// // 	cs devtools.ColorScheme,
// // 	p devtools.Painters,
// // ) *primitives.BoxWidget {

// // 	// Build your main Accountant workspace here.

// // 	_ = b
// // 	_ = p

// // 	return primitives.Box().
// // 		Background(cs.Background)
// // }

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
