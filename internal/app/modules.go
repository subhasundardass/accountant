package app

import (
	appctx "github.com/subhasundardass/accountant/internal/context"
	company "github.com/subhasundardass/accountant/module/company/views"
	ledger "github.com/subhasundardass/accountant/module/ledger/views"
)

func (b *Bootstrap) registerModules() {
	registerFns := []func(*appctx.AppContext){
		company.Register,
		ledger.Register,
		// dashboard.Register,
		// journal.Register,
		// receipt.Register,
		// payment.Register,
		// reports.Register,
		// statement.Register,

	}

	for _, register := range registerFns {
		register(b.AppCtx)
	}
}
