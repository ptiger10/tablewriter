// Package tablewriter provides a basic ASCII table writer
// with customization options for:
// headers,
// label levels,
// cell alignment,
// handling overly-wide cells (truncate vs wrap),
// and auto-merging repeat values in the same column.
package tablewriter

import "io"

// maxColWidth is the max rune width of any column without a header.
// columns with headers have a rune width equal to the widest header.
var maxColWidth int

// A "dividing row" is a row with formatting but no text content.
// Its purpose is to accentuate "content rows".
// There are two types of dividing rows:
// a border, which appears at the top and bottom of the table, and
// a header border, which appears directly below the header rows.
//
// A "content row" is a row with text content.
// Headers, the main body of a table, and footers are all content rows.

var (
	borderEdge,
	borderLabelEdge,
	borderFiller,
	headerEdge,
	headerLabelEdge,
	headerFiller,
	contentEdge,
	contentLabelEdge string
)

// set default values
func resetDefaults() {
	ChangeDefaults(Defaults{
		BorderEdge:       "+",
		BorderLabelEdge:  "++",
		BorderFiller:     "-",
		HeaderEdge:       "|",
		HeaderLabelEdge:  "||",
		HeaderFiller:     "-",
		ContentEdge:      "|",
		ContentLabelEdge: "||",
		MaxColWidth:      30,
	})
}

func init() {
	resetDefaults()
}

// Defaults may be supplied to ChangeDefaults() to change the library's global variable settings.
// All edge and filler symbols must be 1-rune wide, except for label edges which must be 2-runes wide.
// MaxColWidth must be > 0.
// Unsupported field values are ignored.
type Defaults struct {
	BorderEdge, BorderLabelEdge, BorderFiller string
	HeaderEdge, HeaderLabelEdge, HeaderFiller string
	ContentEdge, ContentLabelEdge             string
	MaxColWidth                               int
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
	w                 io.Writer
	rows              [][]string
	alignment         Alignment
	numHeaderRows     int
	numLabelLevels    int
	autoMerge         bool
	truncateCells     bool
	autoCenterHeaders bool
}

func singleWidthString(s string) bool {
	return len([]rune(s)) == 1
}

func doubleWidthString(s string) bool {
	return len([]rune(s)) == 2
}

// ChangeDefaults changes the library's global variable settings for any field supplied.
// Fields with unsupported changes are ignored.
func ChangeDefaults(defaults Defaults) {
	if singleWidthString(defaults.BorderEdge) {
		borderEdge = defaults.BorderEdge
	}
	if doubleWidthString(defaults.BorderLabelEdge) {
		borderLabelEdge = defaults.BorderLabelEdge
	}
	if singleWidthString(defaults.BorderFiller) {
		borderFiller = defaults.BorderFiller
	}
	if singleWidthString(defaults.HeaderEdge) {
		headerEdge = defaults.HeaderEdge
	}
	if doubleWidthString(defaults.HeaderLabelEdge) {
		headerLabelEdge = defaults.HeaderLabelEdge
	}
	if singleWidthString(defaults.HeaderFiller) {
		headerFiller = defaults.HeaderFiller
	}
	if singleWidthString(defaults.ContentEdge) {
		contentEdge = defaults.ContentEdge
	}
	if doubleWidthString(defaults.ContentLabelEdge) {
		contentLabelEdge = defaults.ContentLabelEdge
	}
	if defaults.MaxColWidth > 0 {
		maxColWidth = defaults.MaxColWidth
	}
}
