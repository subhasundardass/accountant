package views

import (
	"github.com/subhasundardass/accountant/internal/context"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/ui"
)

// Register initializes the Ledger module.
func Register(ctx *context.AppContext) {
	if ctx == nil {
		logger.Warn("Ledger module registration skipped: context is nil")
		return
	}

	// ledgerList := NewLedgerList(ctx)
	// _ = ledgerList

	ui.Register("ledger.list", ui.Screen{
		ID:     "ledger.list",
		Title:  "Ledger",
		Render: NewLedgerList,
	})

	ui.Register("group.list", ui.Screen{
		ID:     "group.list",
		Title:  "Ledger",
		Render: NewGroupList,
	})

}
