package context

import (
	"context"
	"sync"

	"github.com/subhasundardass/accountant/ent"
	"github.com/subhasundardass/accountant/internal/config"
	"github.com/subhasundardass/accountant/internal/database"
)

// AppContext contains application-wide runtime state threaded through
// the RetUI render tree.
type AppContext struct {
	Context context.Context

	appName    string
	darkMode   bool
	userName   string
	toggleDark func()

	cfg *config.Config

	DB *database.DB
	mu sync.RWMutex

	// Active company for the current application session.
	companyID int

	// Companies available to the current application session.
	companies []*ent.Company
}

// AppContextValues contains the values used to initialize AppContext.
type AppContextValues struct {
	Context    context.Context
	AppName    string
	DarkMode   bool
	UserName   string
	ToggleDark func()

	DB *database.DB

	// CompanyID is the company currently active in the application.
	CompanyID int

	// Companies available to the current application session.
	Companies []*ent.Company

	// Config is the fully loaded application configuration.
	Config *config.Config
}

// Set initializes the application context.
func (c *AppContext) Set(v AppContextValues) {
	if c == nil {
		return
	}

	c.Context = v.Context
	c.appName = v.AppName
	c.darkMode = v.DarkMode
	c.userName = v.UserName
	c.toggleDark = v.ToggleDark
	c.DB = v.DB
	c.cfg = v.Config

	c.companyID = v.CompanyID
	c.companies = v.Companies

	if v.Config != nil && c.appName == "" {
		c.appName = v.Config.AppName
	}

	if c.userName == "" {
		c.userName = "Guest"
	}
}

// Ctx returns the application context.
func (c *AppContext) Ctx() context.Context {
	if c == nil || c.Context == nil {
		return context.Background()
	}

	return c.Context
}

// CompanyID returns the currently active company ID.
//
// A value of 0 means no company is currently selected.
func (c *AppContext) CompanyID() int {
	if c == nil {
		return 0
	}

	return c.companyID
}

// Companies returns all companies available to the current session.
func (c *AppContext) Companies() []*ent.Company {
	if c == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.companies
}

func (c *AppContext) SetCompanies(companies []*ent.Company) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.companies = companies
}

// Company returns the currently active company.
//
// It returns nil when no company is selected or the selected company
// cannot be found in the current company list.
func (c *AppContext) Company() *ent.Company {
	if c == nil || c.companyID == 0 {
		return nil
	}

	for _, company := range c.companies {
		if company != nil && company.ID == c.companyID {
			return company
		}
	}

	return nil
}

// CompanyName returns the name of the currently active company.
func (c *AppContext) CompanyName() string {
	company := c.Company()

	if company == nil {
		return "Select Company"
	}

	return company.Name
}

// SetCompanyID changes the currently active company.
//
// Persistence of the selected company should be handled by config.Settings;
// this method only changes the runtime application state.
func (c *AppContext) SetCompanyID(id int) {
	if c == nil {
		return
	}

	c.companyID = id
}

// AppName returns the configured application name.
func (c *AppContext) AppName() string {
	if c == nil || c.appName == "" {
		return "App"
	}

	return c.appName
}

// IsDarkMode reports whether dark mode is currently enabled.
func (c *AppContext) IsDarkMode() bool {
	if c == nil {
		return false
	}

	return c.darkMode
}

// UserName returns the current user name.
func (c *AppContext) UserName() string {
	if c == nil || c.userName == "" {
		return "Guest"
	}

	return c.userName
}

// ToggleDark toggles the current dark-mode state.
func (c *AppContext) ToggleDark() {
	if c == nil || c.toggleDark == nil {
		return
	}

	c.toggleDark()
}

// Config returns the fully loaded application configuration.
func (c *AppContext) Config() *config.Config {
	if c == nil {
		return nil
	}

	return c.cfg
}

// Version returns the configured application version.
func (c *AppContext) Version() string {
	if c == nil || c.cfg == nil {
		return ""
	}

	return c.cfg.Version
}

// DateFormat returns the configured display date format.
func (c *AppContext) DateFormat() string {
	if c == nil || c.cfg == nil {
		return "MM/DD/YYYY"
	}

	return c.cfg.DateFormat
}

// Timezone returns the configured timezone.
func (c *AppContext) Timezone() string {
	if c == nil || c.cfg == nil {
		return "UTC"
	}

	return c.cfg.Timezone
}

// Debug reports whether debug mode is enabled.
func (c *AppContext) Debug() bool {
	if c == nil || c.cfg == nil {
		return false
	}

	return c.cfg.Debug
}
