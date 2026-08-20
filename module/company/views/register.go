package views

import (
	"github.com/subhasundardass/accountant/internal/context"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/ui"
)

// Register initializes the Company module.
func Register(ctx *context.AppContext) {
	if ctx == nil {
		logger.Warn("Company module registration skipped: context is nil")
		return
	}

	companyList := NewList(ctx)

	_ = companyList

	ui.Register("company.list", ui.Screen{
		ID:     "company.list",
		Title:  "Companies",
		Render: NewList,
	})

}
