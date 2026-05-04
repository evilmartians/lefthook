package logger

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Color uint8

const (
	ColorCyan Color = iota
	ColorGray
	ColorGreen
	ColorRed
	ColorYellow
	ColorWhite
)

var DefaultColors map[Color]color.Color = map[Color]color.Color{
	ColorCyan:   complete(lipgloss.Color("37"), lipgloss.Color("14"), lipgloss.Color("#70C0BA")),
	ColorGray:   complete(lipgloss.Color("7"), lipgloss.Color("244"), lipgloss.Color("#808080")),
	ColorGreen:  complete(lipgloss.Color("2"), lipgloss.Color("148"), lipgloss.Color("#32cd32")),
	ColorRed:    complete(lipgloss.Color("9"), lipgloss.Color("196"), lipgloss.Color("#ff6347")),
	ColorYellow: complete(lipgloss.Color("11"), lipgloss.Color("191"), lipgloss.Color("#ffaa00")),
}

var NoColors map[Color]color.Color = map[Color]color.Color{
	ColorCyan:   lipgloss.NoColor{},
	ColorGray:   lipgloss.NoColor{},
	ColorGreen:  lipgloss.NoColor{},
	ColorRed:    lipgloss.NoColor{},
	ColorYellow: lipgloss.NoColor{},
}

func (l *Logger) SetColors(colors map[Color]color.Color) {
	l.colors = colors
}

func (l *Logger) Cyan(s string) string {
	return lipgloss.NewStyle().Foreground(l.colors[ColorCyan]).Render(s)
}

func (l *Logger) Gray(s string) string {
	return lipgloss.NewStyle().Foreground(l.colors[ColorGray]).Render(s)
}

func (l *Logger) Yellow(s string) string {
	return lipgloss.NewStyle().Foreground(l.colors[ColorYellow]).Render(s)
}

func (l *Logger) Red(s string) string {
	return lipgloss.NewStyle().Foreground(l.colors[ColorRed]).Render(s)
}

func (l *Logger) Paint(color Color, s string) string {
	return lipgloss.NewStyle().Foreground(l.colors[color]).Render(s)
}
