package views

import (
	"fmt"

	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/subhasundardass/accountant/internal/context"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/module/company"
)

type View struct {
	ctx        *context.AppContext
	controller *company.Controller
	table      *datatable.Widget
}

func NewList(ctx *context.AppContext) widget.Widget {
	if ctx == nil {
		logger.Warn("Company list: context is nil")
		return nil
	}

	v := &View{
		ctx:        ctx,
		controller: company.NewController(ctx),
	}

	companies, err := v.controller.List()
	if err != nil {
		logger.Errorf("Company list: failed to load companies: %v", err)

		return primitives.Box(
			primitives.Text("Failed to load companies"),
		).Padding(16)
	}

	v.table = datatable.New(
		datatable.Columns([]datatable.Column{
			{
				Key:      "id",
				Title:    "ID",
				Width:    80,
				Sortable: true,
			},
			{
				Key:      "name",
				Title:    "Company Name",
				Width:    280,
				Sortable: true,
			},
			{
				Key:      "code",
				Title:    "Code",
				Width:    140,
				Sortable: true,
			},
			{
				Key:      "email",
				Title:    "Email",
				Width:    240,
				Sortable: true,
			},
		}),

		datatable.RowCount(len(companies)),

		datatable.CellValue(func(row int, col string) string {
			if row < 0 || row >= len(companies) {
				return ""
			}

			c := companies[row]

			switch col {
			case "id":
				return fmt.Sprintf("%d", c.ID)
			case "name":
				return c.Name
			case "code":
				return c.Code
			case "email":
				return c.Email
			default:
				return ""
			}
		}),

		datatable.SelectionModeOpt(
			datatable.SelectionSingle,
		),

		datatable.SelectedRow(-1),
	)

	return primitives.Box(
		primitives.Text("Companies").
			FontSize(20).
			Bold(),

		v.table,
	).Padding(16).Gap(12)
}
