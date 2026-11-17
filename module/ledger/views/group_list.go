package views

import (
	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"

	"github.com/subhasundardass/accountant/internal/context"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/module/ledger"
)

type GroupList struct {
	ctx        *context.AppContext
	controller *ledger.Controller
	table      *datatable.Widget
}

func NewGroupList(ctx *context.AppContext) widget.Widget {
	if ctx == nil {
		logger.Warn("Groups list: context is nil")
		return nil
	}

	v := &GroupList{
		ctx:        ctx,
		controller: ledger.NewController(ctx),
	}

	groups, err := v.controller.Groups()
	if err != nil {
		logger.Errorf("Groups list: failed to load groups: %v", err)

		return primitives.Box(
			primitives.Text("Failed to load groups"),
		).Padding(16)
	}

	rows := make([][]string, len(groups))

	for i, rs := range groups {
		rows[i] = []string{
			rs.Code,
			rs.Name,
			rs.Description,
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
				Key:      "description",
				Title:    "Description",
				Width:    400,
				Sortable: false,
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

			case "description":
				return r[2]

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
		primitives.Text("Groups").
			FontSize(20).
			Bold(),

		v.table,
	).Padding(8).Gap(12)
}
