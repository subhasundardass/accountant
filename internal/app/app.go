package app

// import (
// 	"fmt"

// 	"github.com/gogpu/gogpu"
// 	"github.com/gogpu/ui/app"
// 	"github.com/gogpu/ui/desktop"
// 	"github.com/gogpu/ui/primitives"
// 	"github.com/gogpu/ui/theme/devtools"
// 	"github.com/subhasundardass/accountant/internal/logger"
// 	"github.com/subhasundardass/accountant/ui"
// )

// // ============================================================================
// // APPLICATION RUNNER
// // ============================================================================

// // Run initializes and starts the Accountant application.
// func Run() error {

// 	// -------------------------------------------------------------------------
// 	// Bootstrap
// 	// -------------------------------------------------------------------------

// 	b, err := NewBootstrap()
// 	if err != nil {
// 		return fmt.Errorf(
// 			"failed to bootstrap application: %w",
// 			err,
// 		)
// 	}

// 	// Always clean up application resources when the UI exits.
// 	defer func() {
// 		if err := b.Shutdown(); err != nil {
// 			logger.Errorf(
// 				"Application shutdown error: %v",
// 				err,
// 			)
// 		}
// 	}()

// 	// -------------------------------------------------------------------------
// 	// Theme
// 	// -------------------------------------------------------------------------

// 	dt := devtools.NewTheme()
// 	cs := dt.Colors
// 	p := devtools.NewPainters(dt)

// 	// -------------------------------------------------------------------------
// 	// Main window
// 	// -------------------------------------------------------------------------

// 	gogpuApp := gogpu.NewApp(
// 		gogpu.DefaultConfig().
// 			WithTitle("Accountant").
// 			WithSize(800, 600).
// 			WithResizable(true),
// 	)

// 	// Start maximized.
// 	// gogpuApp.Maximize()

// 	// -------------------------------------------------------------------------
// 	// Application workspace
// 	// -------------------------------------------------------------------------

// 	// workspace := buildWorkspace(
// 	// 	b,
// 	// 	cs,
// 	// 	p,
// 	// )

// 	// -------------------------------------------------------------------------
// 	// Root layout
// 	// -------------------------------------------------------------------------

// 	root := ui.Layout(
// 		cs,
// 		p,
// 		// workspace,
// 	)

// 	// -------------------------------------------------------------------------
// 	// gogpu/ui application
// 	// -------------------------------------------------------------------------

// 	uiApp := app.New(
// 		app.WithWindowProvider(gogpuApp),
// 		app.WithPlatformProvider(gogpuApp),
// 		app.WithEventSource(gogpuApp.EventSource()),
// 		// app.WithTheme(dt.AsTheme()),
// 	)

// 	uiApp.SetRoot(root)

// 	// -------------------------------------------------------------------------
// 	// Run event loop
// 	// -------------------------------------------------------------------------

// 	// logger.Info("Accountant UI started")

// 	if err := desktop.Run(gogpuApp, uiApp); err != nil {
// 		return fmt.Errorf(
// 			"UI runtime failed: %w",
// 			err,
// 		)
// 	}

// 	return nil
// }

// // -------------------------------------------------------------------------
// // Workspace
// // -------------------------------------------------------------------------
// func buildWorkspace(
// 	b *Bootstrap,
// 	cs devtools.ColorScheme,
// 	p devtools.Painters,
// ) *primitives.BoxWidget {

// 	// Build your main Accountant workspace here.

// 	_ = b
// 	_ = p

// 	return primitives.Box().
// 		Background(cs.Background)
// }
