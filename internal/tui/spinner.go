package tui

import (
	"os"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

type spinnerModel[R any] struct {
	spinner *spinner.Spinner
	ret     R
	err     error
}

type SpinnerType = spinner.Type

var (
	SpinnerTypeLine    = spinner.Line
	SpinnerTypeDot     = spinner.Dots
	SpinnerTypeMiniDot = spinner.MiniDot
	SpinnerTypeJump    = spinner.Jump
	SpinnerTypePulse   = spinner.Pulse
	SpinnerTypePoints  = spinner.Points
	SpinnerTypeGlobe   = spinner.Globe
	SpinnerTypeMoon    = spinner.Moon
	SpinnerTypeMonkey  = spinner.Monkey
	SpinnerTypeMeter   = spinner.Meter
	SpinnerHamburger   = spinner.Hamburger
	SpinnerEllipsis    = spinner.Ellipsis
)

type SpinnerConfig struct {
	Style      *lipgloss.Style
	Title      string
	TitleStyle *lipgloss.Style
	Type       SpinnerType
}

func NewSpinner[R any](conf SpinnerConfig) *spinnerModel[R] {
	s := spinner.New().Output(os.Stderr)
	if conf.Style != nil {
		s.Style(*conf.Style)
	}
	if conf.Title != "" {
		s.Title(conf.Title)
	}
	if conf.TitleStyle != nil {
		s.Style(*conf.TitleStyle)
	}
	if conf.Type.FPS != 0 {
		s.Type(conf.Type)
	} else {
		s.Type(SpinnerTypePoints)
	}

	return &spinnerModel[R]{
		spinner: s,
	}
}

func (s *spinnerModel[R]) Exec(action func() (R, error)) (R, error) {
	if err := s.spinner.Action(func() {
		s.ret, s.err = action()
	}).Run(); err != nil {
		var zero R
		return zero, err
	}
	return s.ret, s.err
}
