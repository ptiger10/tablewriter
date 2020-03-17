// Package tablewriter provides a basic ASCII table writer
// with customization options for:
// headers,
// label columns,
// cell alignment,
// handling overly-wide cells (truncate vs wrap),
// and auto-merging repeat values in the same column.
package tablewriter

import "io"

// maxColWidth is the max rune width of any column witnout a header.
// columns with headers have a rune width equal to the widest header.
const (
	maxColWidth = 30
)

// A "dividing row" is a row with formatting but no text content.
// Its purpose is to accentuate "content rows".
//
// A "content row" is a row with text content.
// Headers, the main body of a table, and footers are all content rows.

var (
	dividingEdge      = "+"
	dividingLabelEdge = "++"
	dividingFiller    = "-"
	contentEdge       = "|"
	labelEdge         = "||"
)

// SetDividingEdge sets the edge symbol in dividing rows to `symbol` (default: "+")
func SetDividingEdge(symbol string) {
	dividingEdge = symbol
}

// SetDividingLabelEdge sets the edge symbol that demarcates the end of the label levels in dividing rows to `symbol` (default: "++")
func SetDividingLabelEdge(symbol string) {
	dividingLabelEdge = symbol
}

// SetDividingFiller sets the symbol that creates space between dividing edges to `symbol` (default: "-")
func SetDividingFiller(symbol string) {
	dividingFiller = symbol
}

// SetContentEdge sets the edge symbol in content rows to `symbol` (default: "|")
func SetContentEdge(symbol string) {
	contentEdge = symbol
}

// SetLabelEdge sets the edge symbol that demarcates the end of the label levels in content rows to `symbol` (default: "||")
func SetLabelEdge(symbol string) {
	labelEdge = symbol
}

// An Alignment configures how text is aligned in a cell.
type Alignment int

const (
	// AlignCenter centers the cell.
	AlignCenter Alignment = iota
	// AlignRight right-justifies the cell
	AlignRight
	// AlignLeft left-justifies the cell
	AlignLeft
)

// A Table can be rendered into a stringified representation of content rows and dividing rows
// with the results written into an io.Writer.
type Table struct {
	w              io.Writer
	rows           [][]string
	alignment      Alignment
	separateLabels bool
	numHeaderRows  int
	numLabelLevels int
	autoMerge      bool
	truncateCells  bool
}
