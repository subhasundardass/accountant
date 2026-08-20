package ui

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

type workspaceSlot struct {
	widget.WidgetBase
	content widget.Widget
}

var workspace *workspaceSlot

func SetWorkspaceContent(content widget.Widget) {
	if workspace == nil {
		return
	}

	workspace.SetContent(content)
}

func NewWorkspaceSlot() *workspaceSlot {
	s := &workspaceSlot{
		content: primitives.Text("Select a module"),
	}

	workspace = s

	return s
}

func (s *workspaceSlot) SetContent(content widget.Widget) {
	if content == nil {
		return
	}

	s.content = content

	// The entire child has changed.
	s.MarkNeedsLayout()
	s.SetNeedsRedraw(true)
	s.InvalidateScene()
}

func (s *workspaceSlot) Content() widget.Widget {
	return s.content
}

func (s *workspaceSlot) Layout(
	ctx widget.Context,
	c geometry.Constraints,
) geometry.Size {
	if s.content == nil {
		return geometry.Size{}
	}

	return s.content.Layout(ctx, c)
}

func (s *workspaceSlot) Draw(
	ctx widget.Context,
	canvas widget.Canvas,
) {
	if s.content == nil {
		return
	}

	s.content.Draw(ctx, canvas)
}

func (s *workspaceSlot) Event(
	ctx widget.Context,
	e event.Event,
) bool {
	if s.content == nil {
		return false
	}

	return s.content.Event(ctx, e)
}

func (s *workspaceSlot) Children() []widget.Widget {
	if s.content == nil {
		return nil
	}

	return []widget.Widget{s.content}
}
