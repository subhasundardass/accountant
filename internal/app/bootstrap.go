package app

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/subhasundardass/accountant/internal/config"
	appctx "github.com/subhasundardass/accountant/internal/context"
	"github.com/subhasundardass/accountant/internal/database"
	"github.com/subhasundardass/accountant/internal/database/seed"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/module/company"
)

var (
	globalBootstrap *Bootstrap
	bootstrapMu     sync.RWMutex
)

// Bootstrap holds the application's core dependencies and startup state.
type Bootstrap struct {
	DB     *database.DB
	Config *config.Config
	AppCtx *appctx.AppContext

	Hotkeys *GlobalHotkeys

	Ctx    context.Context
	Cancel context.CancelFunc

	cleanup []func() error

	state *appState

	mu sync.RWMutex
}

// ============================================================================
// INITIALIZATION
// ============================================================================

// NewBootstrap creates and initializes the application.
func NewBootstrap() (*Bootstrap, error) {
	logger.Debug("Initializing Bootstrap...")

	// -------------------------------------------------------------------------
	// Configuration
	// -------------------------------------------------------------------------

	cfg := config.Load()
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	b := &Bootstrap{
		Config:  cfg,
		cleanup: make([]func() error, 0),
		Hotkeys: NewGlobalHotkeys(),

		state: &appState{
			config: cfg,
			user: &user{
				name:  "Guest",
				email: "guest@example.com",
				role:  "viewer",
			},
		},
	}

	b.Ctx, b.Cancel = context.WithCancel(context.Background())

	// -------------------------------------------------------------------------
	// Database
	// -------------------------------------------------------------------------

	logger.Debug("Initializing database...")

	db, err := InitDB(cfg)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to initialize database: %w",
			err,
		)
	}

	if db == nil {
		return nil, fmt.Errorf("database client is nil")
	}

	b.DB = db

	b.RegisterCleanup(func() error {
		logger.Debug("Closing database connection...")

		if b.DB != nil {
			return b.DB.Close()
		}

		return nil
	})

	// -------------------------------------------------------------------------
	// Database schema
	// -------------------------------------------------------------------------

	if err := b.runMigrations(); err != nil {
		logger.Warnf("Migration warning: %v", err)
	}

	// -------------------------------------------------------------------------
	// Seed reference data
	// -------------------------------------------------------------------------

	if err := b.seedDatabase(); err != nil {
		logger.Warnf("Seed warning: %v", err)
	}

	// -------------------------------------------------------------------------
	// Application context
	// -------------------------------------------------------------------------

	b.AppCtx = &appctx.AppContext{}

	b.setContext()

	// -------------------------------------------------------------------------
	// Application startup data
	// -------------------------------------------------------------------------

	if err := b.initializeCompany(); err != nil {
		return nil, fmt.Errorf(
			"failed to initialize company: %w",
			err,
		)
	}

	// -------------------------------------------------------------------------
	// Register application modules
	// -------------------------------------------------------------------------

	b.registerModules()
	RegisterApplicationHotkeys(b)

	// -------------------------------------------------------------------------
	// Global bootstrap
	// -------------------------------------------------------------------------

	SetBootstrap(b)

	// retui.Success("Bootstrap completed successfully")

	return b, nil
}

// ============================================================================
// COMPANY INITIALIZATION
// ============================================================================

// initializeCompany loads companies into AppContext and determines the
// active company for the current application session.
//
// Bootstrap is responsible for loading startup data. UI components such as
// CompanySwitcher should only consume the data from AppContext and must not
// query the database during rendering.
func (b *Bootstrap) initializeCompany() error {
	if b.AppCtx == nil {
		return fmt.Errorf("app context is nil")
	}

	// -------------------------------------------------------------------------
	// Load companies
	// -------------------------------------------------------------------------

	controller := company.NewController(b.AppCtx)

	companies, err := controller.List()
	if err != nil {
		return fmt.Errorf(
			"failed to load companies: %w",
			err,
		)
	}

	// -------------------------------------------------------------------------
	// No companies
	// -------------------------------------------------------------------------

	if len(companies) == 0 {
		b.AppCtx.SetCompanyID(0)

		logger.Info(
			"No companies found; company creation required",
		)

		return nil
	}

	// Store the complete company list in application context.

	b.AppCtx.SetCompanies(companies)

	logger.Infof(
		"Loaded %d companies",
		len(companies),
	)

	// -------------------------------------------------------------------------
	// Determine startup company
	// -------------------------------------------------------------------------

	result, err := controller.Initialize()
	if err != nil {
		return fmt.Errorf(
			"failed to initialize company: %w",
			err,
		)
	}

	switch result.State {

	case company.StartupCreateCompany:
		b.AppCtx.SetCompanyID(0)

		logger.Info(
			"No companies found; company creation required",
		)

	case company.StartupSelectCompany:
		// Multiple companies exist but no valid active company has
		// been selected yet.
		b.AppCtx.SetCompanyID(0)

		logger.Info(
			"Multiple companies found; company selection required",
		)

	case company.StartupReady:
		if result.Company == nil {
			return fmt.Errorf(
				"startup company state is ready but company is nil",
			)
		}

		b.AppCtx.SetCompanyID(result.Company.ID)

		logger.Infof(
			"Active company: %s (%d)",
			result.Company.Name,
			result.Company.ID,
		)

	default:
		return fmt.Errorf(
			"unknown company startup state: %d",
			result.State,
		)
	}

	return nil
}

// ============================================================================
// DATABASE FUNCTIONS
// ============================================================================

// InitDB initializes the database connection.
func InitDB(cfg *config.Config) (*database.DB, error) {
	logger.Infof("Connecting to database...")

	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Config.Load() already resolves SQLitePath to an absolute path.
	// Keep that resolved path instead of resolving it again against cwd.
	dbPath := cfg.SQLitePath

	if dbPath == "" {
		dbPath = filepath.Join(
			config.ExecutableDir(),
			"data",
			"retui.db",
		)
	} else if !filepath.IsAbs(dbPath) {
		absPath, err := filepath.Abs(dbPath)
		if err == nil {
			dbPath = absPath
		}
	}

	logger.Infof("Database path: %s", dbPath)

	db, err := database.New("sqlite", dbPath)
	if err != nil {
		logger.Errorf(
			"Failed to open database: %v",
			err,
		)

		return nil, fmt.Errorf(
			"failed to initialize database: %w",
			err,
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	logger.Info("Creating/verifying schema...")

	if err := db.Client.Schema.Create(ctx); err != nil {
		logger.Errorf(
			"Failed to create schema: %v",
			err,
		)

		return nil, fmt.Errorf(
			"failed to create schema: %w",
			err,
		)
	}

	logger.Infof(
		"Database connected: %s",
		dbPath,
	)

	return db, nil
}

// runMigrations runs database migrations.
func (b *Bootstrap) runMigrations() error {
	logger.Info("Running database migrations...")

	ctx, cancel := context.WithTimeout(
		b.Ctx,
		10*time.Second,
	)
	defer cancel()

	if err := b.DB.Client.Schema.Create(ctx); err != nil {
		return fmt.Errorf(
			"failed to create schema: %w",
			err,
		)
	}

	logger.Info("Migrations completed successfully")

	return nil
}

// seedDatabase populates initial reference data.
//
// Seed functions are expected to be safe to execute repeatedly.
func (b *Bootstrap) seedDatabase() error {
	logger.Info("Seeding database...")

	ctx, cancel := context.WithTimeout(
		b.Ctx,
		10*time.Second,
	)
	defer cancel()

	if err := seed.AccountingSeed(
		ctx,
		b.DB.Client,
	); err != nil {
		return fmt.Errorf(
			"failed to seed accounting data: %w",
			err,
		)
	}

	logger.Info("Database seeded successfully")

	return nil
}

// ============================================================================
// CONTEXT FUNCTIONS
// ============================================================================

// setContext initializes the application context.
//
// This initializes stable application dependencies. Dynamic startup data such
// as companies is loaded separately by initializeCompany().
func (b *Bootstrap) setContext() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.AppCtx == nil {
		logger.Error("App context not initialized")
		return
	}

	if b.DB == nil {
		logger.Error("DB is nil in bootstrap")
		return
	}

	userName := "Guest"

	if b.state.user != nil &&
		b.state.user.name != "" {
		userName = b.state.user.name
	}

	b.AppCtx.Set(appctx.AppContextValues{
		Context:   b.Ctx,
		AppName:   b.state.config.AppName,
		DarkMode:  b.state.darkMode,
		UserName:  userName,
		DB:        b.DB,
		Config:    b.state.config,
		CompanyID: b.AppCtx.CompanyID(),

		// Preserve any companies that may already be present.
		Companies: b.AppCtx.Companies(),

		ToggleDark: func() {
			b.mu.Lock()
			b.state.darkMode = !b.state.darkMode
			darkMode := b.state.darkMode
			b.mu.Unlock()

			logger.Debugf(
				"Dark mode toggled: %v",
				darkMode,
			)
		},
	})

	logger.Debug("Context values set successfully")
}

// UpdateContext updates the application context.
func (b *Bootstrap) UpdateContext() {
	b.setContext()
}

// ============================================================================
// CLEANUP AND SHUTDOWN
// ============================================================================

// RegisterCleanup registers a function to execute during shutdown.
func (b *Bootstrap) RegisterCleanup(fn func() error) {
	if fn == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.cleanup = append(
		b.cleanup,
		fn,
	)
}

// Shutdown gracefully shuts down the application.
func (b *Bootstrap) Shutdown() error {
	logger.Debug("Shutting down...")

	// Cancel the application context first so background work can stop.
	if b.Cancel != nil {
		b.Cancel()
	}

	var errs []error

	b.mu.Lock()

	// Execute cleanup in reverse registration order.
	for i := len(b.cleanup) - 1; i >= 0; i-- {
		if err := b.cleanup[i](); err != nil {
			errs = append(errs, err)
		}
	}

	b.mu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors during shutdown: %v",
			errs,
		)
	}

	logger.Info("Shutdown completed successfully")

	return nil
}

// Close is an alias for Shutdown.
func (b *Bootstrap) Close() error {
	return b.Shutdown()
}

// ============================================================================
// GLOBAL BOOTSTRAP
// ============================================================================

// SetBootstrap stores the global bootstrap instance.
func SetBootstrap(b *Bootstrap) {
	bootstrapMu.Lock()
	defer bootstrapMu.Unlock()

	globalBootstrap = b
}

// GetBootstrap returns the global bootstrap instance.
func GetBootstrap() *Bootstrap {
	bootstrapMu.RLock()
	defer bootstrapMu.RUnlock()

	return globalBootstrap
}
