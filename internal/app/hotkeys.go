package app

import (
	"strings"
	"sync"

	"github.com/gogpu/gpucontext"
	"github.com/subhasundardass/accountant/internal/logger"
	"github.com/subhasundardass/accountant/ui"
)

// HotkeyAction is called when a registered hotkey is triggered.
type HotkeyAction func()

// GlobalHotkeys manages application-wide keyboard shortcuts.
type GlobalHotkeys struct {
	mu      sync.RWMutex
	actions map[string]HotkeyAction
}

// NewGlobalHotkeys creates a global hotkey manager.
func NewGlobalHotkeys() *GlobalHotkeys {
	return &GlobalHotkeys{
		actions: make(map[string]HotkeyAction),
	}
}

// Register registers an application-wide hotkey.
func (h *GlobalHotkeys) Register(key string, action HotkeyAction) {
	if h == nil || action == nil {
		return
	}

	key = normalizeHotkey(key)

	h.mu.Lock()
	h.actions[key] = action
	h.mu.Unlock()

	logger.Debugf("Global hotkey registered: %s", key)
}

// Unregister removes a hotkey.
func (h *GlobalHotkeys) Unregister(key string) {
	if h == nil {
		return
	}

	key = normalizeHotkey(key)

	h.mu.Lock()
	delete(h.actions, key)
	h.mu.Unlock()

	logger.Debugf("Global hotkey unregistered: %s", key)
}

// Trigger executes the action registered for a hotkey.
func (h *GlobalHotkeys) Trigger(key string) bool {
	if h == nil {
		return false
	}

	key = normalizeHotkey(key)

	h.mu.RLock()
	action, ok := h.actions[key]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	action()
	return true
}

// Has reports whether a hotkey is registered.
func (h *GlobalHotkeys) Has(key string) bool {
	if h == nil {
		return false
	}

	key = normalizeHotkey(key)

	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.actions[key]
	return ok
}

// Attach connects the hotkey manager to the GoGPU keyboard event source.
//
// This is the bridge between GoGPU keyboard events and application actions.
func (h *GlobalHotkeys) Attach(eventSource interface {
	OnKeyPress(func(gpucontext.Key, gpucontext.Modifiers))
}) {
	if h == nil || eventSource == nil {
		return
	}

	eventSource.OnKeyPress(func(
		key gpucontext.Key,
		mods gpucontext.Modifiers,
	) {
		hotkey := makeHotkey(key, mods)

		if h.Trigger(hotkey) {
			logger.Debugf("Global hotkey triggered: %s", hotkey)
		}
	})
}

func (h *GlobalHotkeys) Handle(
	key gpucontext.Key,
	mods gpucontext.Modifiers,
) {
	if h == nil {
		return
	}

	hotkey := makeHotkey(key, mods)

	h.Trigger(hotkey)
}

// RegisterApplicationHotkeys registers the application's standard shortcuts.
func RegisterApplicationHotkeys(b *Bootstrap) {
	if b == nil || b.Hotkeys == nil {
		return
	}

	h := b.Hotkeys

	h.Register("F1", func() {
		logger.Info("F1 - Select Company requested")
		ui.OpenScreen("company.list", b.AppCtx)
	})

	h.Register("F2", func() {
		logger.Info("Accounts requested")
	})

	h.Register("F6", func() {
		logger.Info("Ledger requested")
	})

	h.Register("F10", func() {
		logger.Info("Search requested")
	})

	h.Register("F12", func() {
		logger.Info("Settings requested")
	})

	logger.Debug("Application hotkeys registered")
}

// normalizeHotkey makes registrations consistent.
func normalizeHotkey(key string) string {
	return strings.ToUpper(strings.TrimSpace(key))
}

// makeHotkey converts GoGPU keyboard events into our application hotkey name.
func makeHotkey(
	key gpucontext.Key,
	mods gpucontext.Modifiers,
) string {
	// Start with modifiers.
	var parts []string

	// The exact modifier constants depend on your gpucontext version.
	// Add the appropriate checks here for Ctrl/Alt/Shift.
	//
	// Example:
	//
	// if mods&gpucontext.ModifierCtrl != 0 {
	//     parts = append(parts, "CTRL")
	// }

	parts = append(parts, keyName(key))

	return strings.Join(parts, "+")
}

// keyName converts a GoGPU key into our application key name.
func keyName(key gpucontext.Key) string {
	switch key {
	case gpucontext.KeyF1:
		return "F1"
	case gpucontext.KeyF2:
		return "F2"
	case gpucontext.KeyF6:
		return "F6"
	case gpucontext.KeyF10:
		return "F10"
	case gpucontext.KeyF12:
		return "F12"
	default:
		return strings.ToUpper(key.String())
	}
}
