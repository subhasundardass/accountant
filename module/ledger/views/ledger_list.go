package views

import (
	"fmt"

	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/subhasundardass/accountant/internal/context"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/module/ledger"
)

type LedgerList struct {
	ctx        *context.AppContext
	controller *ledger.Controller
	table      *datatable.Widget
}

func NewLedgerList(ctx *context.AppContext) widget.Widget {

	v := &LedgerList{
		controller: ledger.NewController(ctx),
	}

	ledgers, err := v.controller.Ledgers(0)
	if err != nil {
		logger.Errorf("Ledger list: failed to load companies: %v", err)

		return primitives.Box(
			primitives.Text("Failed to load companies"),
		).Padding(16)
	}

	rows := make([][]string, len(ledgers))
	for i, rs := range ledgers {
		groupName := ""
		if rs.Edges.Group != nil {
			groupName = rs.Edges.Group.Name
		}

		rows[i] = []string{
			rs.Code,
			rs.Name,
			groupName, // Display group name instead of ID
			rs.Description,
			fmt.Sprintf("%.2f", rs.Balance),
		}
	}

	v.table = datatable.New(
		datatable.Columns([]datatable.Column{
			{
				Key:      "code",
				Title:    "Code",
				Width:    140,
				Sortable: true,
			},
			{
				Key:      "name",
				Title:    "Name",
				Width:    300,
				Sortable: true,
			},

			{
				Key:      "group",
				Title:    "Group",
				Width:    300,
				Sortable: true,
			},
			{
				Key:      "description",
				Title:    "Description",
				Width:    400,
				Sortable: false,
			},
			{
				Key:      "balance",
				Title:    "Balance",
				Width:    200,
				Sortable: true,
				Align:    widget.TextAlignRight,
			},
		}),

		datatable.RowCount(len(rows)),

		datatable.CellValue(func(row int, col string) string {
			if row < 0 || row >= len(rows) {
				return ""
			}

			r := rows[row]

			switch col {
			case "code":
				return r[0]

			case "name":
				return r[1]

			case "group":
				return r[2]

			case "description":
				return r[3]

			case "balance":
				return r[4]
			default:
				return ""
			}
		}),

		datatable.SelectionModeOpt(
			datatable.SelectionSingle,
		),

		datatable.SelectedRow(0),
	)

	return primitives.Box(
		primitives.Text("Ledgers").
			FontSize(20).
			Bold(),

		v.table,
	).Padding(8).Gap(12)
}
