# Termyx

**Termyx** is a declarative terminal UI runtime for Go — think Ink for the terminal, but in Go.

Build beautiful, dynamic terminal applications using a component tree, flexbox layout, a reconciler, and a diff-based cell renderer. No manual cursor positioning. No string juggling. Just declare your UI and let Termyx handle the rest.

```go
import termyx "github.com/stackxio/termyx"

termyx.Run(&termyx.App{
    Root: func() *termyx.Node {
        return termyx.Row(
            termyx.Border(" Logs ", borderStyle, logPane),
            termyx.Column(
                termyx.Border(" Metrics ", borderStyle, metricsPane),
                termyx.Border(" Status ", borderStyle, statusPane),
            ),
        )
    },
    OnKey: func(ev termyx.KeyEvent) bool {
        return ev.Rune == 'q' || ev.Key == termyx.KeyCtrlC
    },
})
```

---

## Install

```bash
go get github.com/stackxio/termyx@latest
```

---

## How it works

Every frame follows this pipeline:

```
App State
  → Root() produces a Node tree
  → Reconciler diffs with the previous tree (stable IDs, keyed children)
  → Layout engine runs a flexbox pass (Row/Column, padding, min-size, alignment)
  → Render pass fills a cell Buffer (rune + Style per cell)
  → Diff Painter writes only the changed cells as ANSI escape sequences
  → Terminal
```

---

## Core API

### Layout

```go
termyx.Row(children...)        // horizontal flex container
termyx.Column(children...)     // vertical flex container
termyx.Grow(node, 2.0)         // flex grow factor
termyx.Fixed(node, w, h)       // fixed width + height
termyx.FixedWidth(node, w)     // fixed width only
termyx.FixedHeight(node, h)    // fixed height only
termyx.Pad(node, 1)            // uniform padding
termyx.PadXY(node, x, y)       // horizontal + vertical padding
termyx.PadSides(node, t,r,b,l) // per-side padding
termyx.MinSize(node, w, h)     // minimum dimensions
termyx.WithAlign(node, termyx.AlignCenter) // cross-axis alignment
```

### Content

```go
termyx.Text("hello")
termyx.StyledText("hello", style)
termyx.Custom(func(buf *termyx.Buffer, l termyx.LayoutResult) {
    c := buf.Region(l)        // Canvas with relative coords
    c.Fill(0, 0, c.Width, c.Height, ' ', bg)
    c.WriteText(2, 1, "hello", titleStyle)
    c.WriteTextTrunc(2, 2, 20, longString, dimStyle)
})
termyx.Border(" Title ", borderStyle, child)
```

### Style

```go
termyx.Style{
    FG:            termyx.RGB(30, 215, 96),
    BG:            termyx.Hex("#0d1117"),
    Bold:          true,
    Italic:        true,
    Underline:     true,
    Strikethrough: true,
    Reverse:       true, // swap FG/BG — used for cursor/selection
}
```

### Scroll

```go
var scroll termyx.ScrollState

// In OnKey:
case termyx.KeyDown: scroll.ScrollDown(1)
case termyx.KeyUp:   scroll.ScrollUp(1)
case termyx.KeyEnd:  scroll.ScrollToBottom()

// In Root:
termyx.Scroll(&scroll, contentNode, 0)
```

### Focus

```go
// Mark a node as focusable with an OnKey handler:
termyx.Focusable(node, func(ev termyx.KeyEvent) {
    // handle keys when this pane is focused
})

// Tab / Shift+Tab cycle focus automatically.
// node.Focused is true when the node is active — use it to style the border.
```

### Overlay / Modal

```go
termyx.Overlay(base, modal, isVisible, termyx.OverlayOpts{
    Width: 60, Height: 20,
})
termyx.Backdrop(dimStyle, modal) // fill region before rendering modal
```

---

## Built-in Components

### List

```go
model := &termyx.ListModel{
    Items: []termyx.ListItem{
        {Label: "pod-abc", Style: theme.StyleSuccess()},
        {Label: "pod-xyz"},
    },
    Selected: 0,
    Style:    theme.StyleList(),
}
// In OnKey: model.SelectNext() / model.SelectPrev()
termyx.List(model)
```

### Table

```go
model := &termyx.TableModel{
    Columns: []termyx.TableColumn{
        {Title: "NAME", Width: 30},
        {Title: "STATUS"},
        {Title: "AGE", Width: 8},
    },
    Rows: []termyx.TableRow{
        {Cells: []string{"my-pod", "Running", "2d"}},
    },
    Selected: 0,
    Style:    theme.StyleTable(),
    Border:   true,
}
termyx.Table(model)
```

### TabBar

```go
termyx.TabBar([]string{"Pods", "Nodes", "Events"}, activeTab, theme.StyleTabBar())
```

### StatusBar

```go
termyx.FixedHeight(
    termyx.StatusBar("cluster: prod", "q quit  ? help", statusStyle),
    1,
)
```

### Spinner

```go
spinner := &termyx.SpinnerModel{Label: "Loading...", Style: dimStyle, Active: true}
spinner.StartAutoTick(80*time.Millisecond, updateCh, stopCh)

// In Root:
termyx.Spinner(spinner)
```

### TextInput

```go
input := &termyx.TextInputModel{Placeholder: "Search...", Style: inputStyle}

// In OnKey (when search mode is active):
input.HandleKey(ev)

// In Root:
termyx.TextInput(input)
```

### ProgressBar

```go
termyx.ProgressBar(cpu, " CPU", filledStyle, emptyStyle)
```

---

## Themes

```go
// Pick a theme — swapping updates every derived style.
theme := termyx.DarkTheme    // or LightTheme, DraculaTheme, TokyoNightTheme

// Derive styles:
theme.StyleNormal()
theme.StyleSelected()
theme.StyleHeader()
theme.StyleSuccess()
theme.StyleError()
theme.StyleBorder()
theme.StyleActiveBorder()
theme.StyleTabBar()   // → TabStyle
theme.StyleList()     // → ListStyle
theme.StyleTable()    // → TableStyle
```

---

## Text utilities

```go
termyx.Truncate("long string", 10)      // → "long str…"
termyx.TruncateHard("long string", 10)  // → "long strin"
termyx.PadRight("hi", 10)              // → "hi        "
termyx.PadLeft("hi", 10)               // → "        hi"
termyx.Center("hi", 10)               // → "    hi    "
termyx.WordWrap("long text here", 8)   // → []string{"long", "text", "here"}
```

---

## Canvas (relative-coordinate drawing)

Custom nodes get a `*Buffer` + `LayoutResult`. Call `buf.Region(layout)` to get a `Canvas` that works in local (0, 0)-relative coordinates — no manual `l.X + offset` arithmetic:

```go
termyx.Custom(func(buf *termyx.Buffer, l termyx.LayoutResult) {
    c := buf.Region(l)
    c.Fill(0, 0, c.Width, c.Height, ' ', bg)         // fill background
    c.WriteText(2, 0, "Title", titleStyle)            // local coords
    c.HLine(1, '─', borderStyle)                     // full-width divider
    sub := c.Sub(2, 2, c.Width-4, c.Height-4)        // sub-region
    sub.WriteText(0, 0, content, normalStyle)
})
```

---

## Running the dashboard example

```bash
git clone https://github.com/stackxio/termyx
cd termyx
go run ./examples/dashboard/
# Press q or Ctrl+C to exit
```

---

## Versioning

A new semver release is created automatically on every push to `main`. Install a specific version:

```bash
go get github.com/stackxio/termyx@v0.1.3
```

---

## Roadmap

- [ ] Mouse support (click, wheel)
- [ ] Wide character / CJK width support
- [ ] Grid layout node (for heatmap-style views)
- [ ] Node render caching (skip re-render if inputs unchanged)
- [ ] Vaxis backend option

---

## License

MIT
