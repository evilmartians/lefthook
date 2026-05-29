package logger

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

type Color uint8

const (
	ColorWhite Color = iota
	ColorCyan
	ColorGray
	ColorGreen
	ColorRed
	ColorYellow
)

type colorsFlag int

const (
	colorsEnabled colorsFlag = iota
	colorsDisabled
	colorsCustom
)

type ColorsSetting struct {
	kind   colorsFlag
	colors map[Color]color.Color
}

func (c ColorsSetting) get(color Color) color.Color {
	return c.colors[color]
}

var (
	profile  = colorprofile.Detect(os.Stdout, os.Environ())
	complete = lipgloss.Complete(profile)
)

var DefaultColors ColorsSetting = ColorsSetting{
	kind: colorsEnabled,
	colors: map[Color]color.Color{
		ColorCyan:   complete(lipgloss.Color("14"), lipgloss.Color("73"), lipgloss.Color("#5FAFAF")),
		ColorGray:   complete(lipgloss.Color("8"), lipgloss.Color("102"), lipgloss.Color("#878787")),
		ColorGreen:  complete(lipgloss.Color("2"), lipgloss.Color("34"), lipgloss.Color("#00AF00")),
		ColorRed:    complete(lipgloss.Color("9"), lipgloss.Color("203"), lipgloss.Color("#FF5F5F")),
		ColorYellow: complete(lipgloss.Color("3"), lipgloss.Color("3"), lipgloss.Color("#808000")),
		ColorWhite:  complete(lipgloss.Color("15"), lipgloss.Color("15"), lipgloss.Color("#FFFFFF")),
	},
}

var NoColors ColorsSetting = ColorsSetting{
	kind: colorsDisabled,
	colors: map[Color]color.Color{
		ColorCyan:   lipgloss.NoColor{},
		ColorGray:   lipgloss.NoColor{},
		ColorGreen:  lipgloss.NoColor{},
		ColorRed:    lipgloss.NoColor{},
		ColorYellow: lipgloss.NoColor{},
		ColorWhite:  lipgloss.NoColor{},
	},
}

func (l *Logger) SetColors(colors map[Color]color.Color) {
	if l.colorsForced {
		return
	}

	for color, existingValue := range l.colors.colors {
		if _, ok := colors[color]; !ok {
			colors[color] = existingValue
		}
	}

	l.colors = ColorsSetting{
		kind:   colorsCustom,
		colors: colors,
	}
	l.colorsForced = true
}

func (l *Logger) EnableColors() {
	if l.colorsForced {
		return
	}

	l.colors = DefaultColors
	l.colorsForced = true
}

func (l *Logger) DisableColors() {
	if l.colorsForced {
		return
	}

	l.colors = NoColors
	l.colorsForced = true
}

func (l *Logger) Paint(color Color, s string) string {
	return lipgloss.NewStyle().Foreground(l.colors.get(color)).Render(s)
}
