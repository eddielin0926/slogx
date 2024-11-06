package slogx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sync"
)

type IndentHandler struct {
	opts           Options
	preformatted   []byte   // data from WithGroup and WithAttrs
	unopenedGroups []string // groups from WithGroup that haven't been opened
	indentLevel    int      // same as number of opened groups so far
	mu             *sync.Mutex
	out            io.Writer
}

type Options struct {
	// Level reports the minimum level to log.
	// Levels with lower levels are discarded.
	// If nil, the Handler uses [slog.LevelInfo].
	Level slog.Leveler
}

func New(out io.Writer, opts *Options) *IndentHandler {
	h := &IndentHandler{out: out, mu: &sync.Mutex{}}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return h
}

func (h *IndentHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *IndentHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)
	if !r.Time.IsZero() {
		buf = h.appendAttr(buf, slog.Time(slog.TimeKey, r.Time), 0)
	}

	// Insert preformatted attributes just after built-in ones.
	buf = append(buf, h.preformatted...) // FIXME: weird place to put this

	// Extract the "front" attribute if it exists and format it.
	var frontValue string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "front" {
			frontValue = a.Value.String()
			return false // Stop iteration once "front" is found.
		}
		return true
	})

	// Append the "front" key value at the beginning.
	if frontValue != "" {
		buf = h.appendAttr(buf, slog.Any("front", frontValue), 0)
	}
	buf = h.appendAttr(buf, slog.Any(slog.LevelKey, r.Level), 0)
	buf = h.appendAttr(buf, slog.String(slog.MessageKey, r.Message), 0)
	// if r.PC != 0 {
	// 	fs := runtime.CallersFrames([]uintptr{r.PC})
	// 	f, _ := fs.Next()
	// 	buf = h.appendAttr(buf, slog.String(slog.SourceKey, fmt.Sprintf("(%s:%d)", f.File, f.Line)), 0)
	// }
	buf = append(buf, '\n')

	if r.NumAttrs() > 0 {
		buf = h.appendUnopenedGroups(buf, h.indentLevel)
		r.Attrs(func(a slog.Attr) bool {
			buf = append(buf, '\t')
			buf = h.appendAttr(buf, a, h.indentLevel+len(h.unopenedGroups))
			buf = append(buf, '\n')
			return true
		})
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf)
	return err
}

func (h *IndentHandler) appendAttr(buf []byte, a slog.Attr, indentLevel int) []byte {
	// Resolve the Attr's value before doing anything else.
	a.Value = a.Value.Resolve()
	// Ignore empty Attrs.
	if a.Equal(slog.Attr{}) {
		return buf
	}
	if len(buf) > 0 { // FIXME: Not adding space with key: front
		buf = append(buf, ' ')
	}
	switch a.Value.Kind() {
	case slog.KindString:
		if a.Key == "front" || a.Key == "source" || a.Key == "msg" {
			buf = fmt.Appendf(buf, "%s", a.Value.String())
		} else {
			buf = fmt.Appendf(buf, colorize(fmt.Sprintf("%s: \"%v\"", a.Key, a.Value), Gray))
		}
	case slog.KindTime:
		// Write times in a standard way, without the monotonic time.
		buf = fmt.Appendf(buf, colorize(a.Value.Time().Format("15:04:05.000"), Blue))
	case slog.KindGroup:
		attrs := a.Value.Group()
		// Ignore empty groups.
		if len(attrs) == 0 {
			return buf
		}
		// If the key is non-empty, write it out and indent the rest of the attrs.
		// Otherwise, inline the attrs.
		if a.Key != "" {
			buf = fmt.Appendf(buf, "%s", a.Key)
			indentLevel++
		}
		for _, ga := range attrs {
			buf = h.appendAttr(buf, ga, indentLevel)
		}
	default:
		if a.Key == "level" {
			switch a.Value.String() {
			case "DEBUG":
				buf = fmt.Appendf(buf, "%-14s", colorize(a.Value.String(), LightCyan))
			case "INFO":
				buf = fmt.Appendf(buf, "%-14s", colorize(a.Value.String(), Green))
			case "WARN":
				buf = fmt.Appendf(buf, "%-14s", colorize(a.Value.String(), Yellow))
			case "ERROR":
				buf = fmt.Appendf(buf, "%-14s", colorize(a.Value.String(), LightRed))
			default:
				buf = fmt.Appendf(buf, "%-14s", a.Value)
			}
		} else {
			buf = fmt.Appendf(buf, colorize(fmt.Sprintf("%s: %v", a.Key, a.Value), Gray))
		}
	}
	return buf
}

func (h *IndentHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	// Add an unopened group to h2 without modifying h.
	h2.unopenedGroups = make([]string, len(h.unopenedGroups)+1)
	copy(h2.unopenedGroups, h.unopenedGroups)
	h2.unopenedGroups[len(h2.unopenedGroups)-1] = name
	return &h2
}

func (h *IndentHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := *h
	// Force an append to copy the underlying array.
	pre := slices.Clip(h.preformatted)
	// Add all groups from WithGroup that haven't already been added.
	h2.preformatted = h2.appendUnopenedGroups(pre, h2.indentLevel)
	// Each of those groups increased the indent level by 1.
	h2.indentLevel += len(h2.unopenedGroups)
	// Now all groups have been opened.
	h2.unopenedGroups = nil
	// Pre-format the attributes.
	for _, a := range attrs {
		h2.preformatted = h2.appendAttr(h2.preformatted, a, h2.indentLevel)
	}
	return &h2
}

func (h *IndentHandler) appendUnopenedGroups(buf []byte, indentLevel int) []byte {
	for _, g := range h.unopenedGroups {
		buf = fmt.Appendf(buf, "%*s%s:\n", indentLevel*4, "", g)
		indentLevel++
	}
	return buf
}

func colorize(s string, color string) string {
	return color + s + Reset
}
