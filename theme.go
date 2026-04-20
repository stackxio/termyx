package termyx

// Theme is a named palette of colors. Applications define one theme and derive
// all component styles from it, enabling dark/light switching by swapping themes.
type Theme struct {
	Bg       Color // primary background
	BgAlt    Color // alternate background (panels, sidebars)
	BgSel    Color // selection/highlight background
	Fg       Color // primary foreground
	FgDim    Color // dimmed foreground (hints, timestamps)
	FgBright Color // bright foreground (titles, emphasis)
	Accent   Color // accent (active tab, focused border)
	Success  Color // green (running, ready)
	Warning  Color // yellow (pending, degraded)
	Error    Color // red (failed, critical)
	Info     Color // blue/cyan (informational)
	Border   Color // border lines
}

// Style helpers — derive ready-to-use Style values from the theme.

func (t Theme) StyleNormal() Style      { return Style{FG: t.Fg, BG: t.Bg} }
func (t Theme) StyleDim() Style         { return Style{FG: t.FgDim, BG: t.Bg} }
func (t Theme) StyleBold() Style        { return Style{FG: t.FgBright, BG: t.Bg, Bold: true} }
func (t Theme) StyleSelected() Style    { return Style{FG: t.FgBright, BG: t.BgSel, Bold: true} }
func (t Theme) StyleHeader() Style      { return Style{FG: t.Accent, BG: t.Bg, Bold: true} }
func (t Theme) StyleBorder() Style      { return Style{FG: t.Border, BG: t.Bg} }
func (t Theme) StyleActiveBorder() Style { return Style{FG: t.Accent, BG: t.Bg} }
func (t Theme) StyleSuccess() Style     { return Style{FG: t.Success, BG: t.Bg} }
func (t Theme) StyleWarning() Style     { return Style{FG: t.Warning, BG: t.Bg} }
func (t Theme) StyleError() Style       { return Style{FG: t.Error, BG: t.Bg} }
func (t Theme) StyleInfo() Style        { return Style{FG: t.Info, BG: t.Bg} }
func (t Theme) StyleAlt() Style         { return Style{FG: t.Fg, BG: t.BgAlt} }

func (t Theme) StyleTabBar() TabStyle {
	return TabStyle{
		Active:   Style{FG: t.Bg, BG: t.Accent, Bold: true},
		Inactive: Style{FG: t.FgDim, BG: t.BgAlt},
		Gap:      Style{FG: t.Fg, BG: t.Bg},
	}
}

func (t Theme) StyleList() ListStyle {
	return ListStyle{
		Normal:   t.StyleNormal(),
		Selected: t.StyleSelected(),
		Header:   t.StyleHeader(),
	}
}

func (t Theme) StyleTable() TableStyle {
	return TableStyle{
		Header:   t.StyleHeader(),
		Normal:   t.StyleNormal(),
		Selected: t.StyleSelected(),
		Border:   t.StyleBorder(),
	}
}

// ── Built-in themes ───────────────────────────────────────────────────────────

// DarkTheme is a GitHub-inspired dark theme.
var DarkTheme = Theme{
	Bg:       Hex("#0d1117"),
	BgAlt:    Hex("#161b22"),
	BgSel:    Hex("#1f3a5f"),
	Fg:       Hex("#c9d1d9"),
	FgDim:    Hex("#6e7681"),
	FgBright: Hex("#f0f6fc"),
	Accent:   Hex("#58a6ff"),
	Success:  Hex("#3fb950"),
	Warning:  Hex("#d29922"),
	Error:    Hex("#f85149"),
	Info:     Hex("#79c0ff"),
	Border:   Hex("#30363d"),
}

// LightTheme is a GitHub-inspired light theme.
var LightTheme = Theme{
	Bg:       Hex("#ffffff"),
	BgAlt:    Hex("#f6f8fa"),
	BgSel:    Hex("#dbeafe"),
	Fg:       Hex("#24292f"),
	FgDim:    Hex("#57606a"),
	FgBright: Hex("#0d1117"),
	Accent:   Hex("#0969da"),
	Success:  Hex("#1a7f37"),
	Warning:  Hex("#9a6700"),
	Error:    Hex("#cf222e"),
	Info:     Hex("#0550ae"),
	Border:   Hex("#d0d7de"),
}

// DraculaTheme uses the popular Dracula color scheme.
var DraculaTheme = Theme{
	Bg:       Hex("#282a36"),
	BgAlt:    Hex("#21222c"),
	BgSel:    Hex("#44475a"),
	Fg:       Hex("#f8f8f2"),
	FgDim:    Hex("#6272a4"),
	FgBright: Hex("#ffffff"),
	Accent:   Hex("#bd93f9"),
	Success:  Hex("#50fa7b"),
	Warning:  Hex("#ffb86c"),
	Error:    Hex("#ff5555"),
	Info:     Hex("#8be9fd"),
	Border:   Hex("#44475a"),
}

// TokyoNightTheme uses the Tokyo Night color scheme popular in editors.
var TokyoNightTheme = Theme{
	Bg:       Hex("#1a1b26"),
	BgAlt:    Hex("#16161e"),
	BgSel:    Hex("#283457"),
	Fg:       Hex("#a9b1d6"),
	FgDim:    Hex("#565f89"),
	FgBright: Hex("#c0caf5"),
	Accent:   Hex("#7aa2f7"),
	Success:  Hex("#9ece6a"),
	Warning:  Hex("#e0af68"),
	Error:    Hex("#f7768e"),
	Info:     Hex("#7dcfff"),
	Border:   Hex("#292e42"),
}
