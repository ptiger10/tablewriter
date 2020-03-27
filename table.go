package tablewriter

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// NewTable creates a default table writing to `w`.
func NewTable(w io.Writer) *Table {
	return &Table{
		w:                 w,
		rows:              [][]string{},
		alignment:         AlignCenter,
		numHeaderRows:     0,
		numLabelLevels:    0,
		autoMerge:         false,
		truncateCells:     false,
		autoCenterHeaders: true,
	}
}

func (tbl *Table) sameShape(row []string) error {
	// no rows in table? ok
	if len(tbl.rows) == 0 {
		return nil
	}
	// shape does not match? bad
	if len(row) != len(tbl.rows[0]) {
		return fmt.Errorf("new row must have same number of fields as all existing rows in Table (%d != %d)", len(row), len(tbl.rows[0]))
	}
	// shape matches? ok
	return nil
}

// AppendHeaderRow appends a header row to the table.
func (tbl *Table) AppendHeaderRow(row []string) error {
	err := tbl.sameShape(row)
	if err != nil {
		return fmt.Errorf("appending header row: %v", err)
	}

	headersOnly := make([][]string, tbl.numHeaderRows)
	copy(headersOnly, tbl.rows[:tbl.numHeaderRows])
	headersOnly = append(headersOnly, row)

	tbl.rows = append(headersOnly, tbl.rows[tbl.numHeaderRows:]...)
	tbl.numHeaderRows++
	return nil
}

// AppendRow appends a non-header row to the table.
func (tbl *Table) AppendRow(row []string) error {
	err := tbl.sameShape(row)
	if err != nil {
		return fmt.Errorf("appending row (%v): %v", row, err)
	}
	tbl.rows = append(tbl.rows, row)
	return nil
}

// AppendRows appends one or more non-header rows to the table.
func (tbl *Table) AppendRows(rows [][]string) error {
	for i := range rows {
		err := tbl.AppendRow(rows[i])
		if err != nil {
			return fmt.Errorf("appending rows: position %d: %v", i, err)
		}
	}
	return nil
}

// DisableHeaderAutoCentering causes header cells to be aligned based on the underlying table alignment (default: headers are auto-centered).
func (tbl *Table) DisableHeaderAutoCentering() {
	tbl.autoCenterHeaders = false
}

// MergeRepeats merges all repeated values in a column together.
func (tbl *Table) MergeRepeats() {
	tbl.autoMerge = true
}

// TruncateWideCells handles overly wide cells by truncating them (default: wrap cell remainder onto new one or more new lines).
func (tbl *Table) TruncateWideCells() {
	tbl.truncateCells = true
}

// SetAlignment sets the alignment of cells in content rows to `alignment`.
func (tbl *Table) SetAlignment(alignment Alignment) {
	tbl.alignment = alignment
}

// SetLabelLevelCount sets the number of label levels to `n`.
// "Label levels" are the leftmost columns in the table, and typically have values that help identify ("label") specific rows.
// They are often analogous to a table index.
// Setting `n` > 0 will visually separate the label levels from the other columns in the table.
// (Default: 0 label levels).
func (tbl *Table) SetLabelLevelCount(n int) {
	tbl.numLabelLevels = n
}

// creates a stringified representation of content rows and dividing rows
func (tbl *Table) render() (string, error) {
	if len(tbl.rows) == 0 {
		return "", fmt.Errorf("table must have at least 1 row")
	}
	colWidths := tbl.resizeColWidths()
	borderLine := stringifyDividingRow(colWidths, tbl.numLabelLevels, false)
	headerLine := stringifyDividingRow(colWidths, tbl.numLabelLevels, true)

	var ret string
	var priorRow []string
	for i := range tbl.rows {
		// write a borderLine at the top and a headerLine after the last header row
		if i == 0 {
			ret += borderLine
		} else if i == tbl.numHeaderRows {
			ret += headerLine
		}
		// copy row to avoid changing original in calls to autoMergeRows and stringifyContentRow
		rowCopy := make([]string, len(tbl.rows[i]))
		copy(rowCopy, tbl.rows[i])
		if tbl.autoMerge {
			// auto-merge applies only to non-header rows
			if i == tbl.numHeaderRows+1 {
				priorRow = tbl.rows[tbl.numHeaderRows]
			}
			autoMergeRows(priorRow, rowCopy)
		}
		isHeader := i < tbl.numHeaderRows
		ret += tbl.stringifyContentRow(colWidths, rowCopy, isHeader)
	}
	// write a borderLine at the bottom
	ret += borderLine
	return ret, nil
}

// Render creates a stringified representation of content rows and dividing rows
// and writes the results into the table's io.Writer.
func (tbl *Table) Render() error {
	s, err := tbl.render()
	if err != nil {
		return fmt.Errorf("tbl.Render(): %v", err)
	}
	_, err = tbl.w.Write([]byte(s))
	if err != nil {
		return fmt.Errorf("tbl.Render(): %v", err)
	}
	return nil
}

// modify priorRow and currentRow in place
func autoMergeRows(priorRow, currentRow []string) {
	for k := range priorRow {
		if priorRow[k] == currentRow[k] {
			currentRow[k] = ""
		} else {
			priorRow[k] = currentRow[k]
		}
	}
}

func runeWidth(s string) int {
	return len([]rune(s))
}

// expects all rows to have the same number of columns
// expects len(tbl.rows) to be greater than 0.
func (tbl *Table) resizeColWidths() []int {
	ret := make([]int, len(tbl.rows[0]))
	for i := range tbl.rows {
		for k := range tbl.rows[i] {
			// header row? column width may exceed max width
			if i < tbl.numHeaderRows {
				if headerWidth := runeWidth(tbl.rows[i][k]); headerWidth > ret[k] {
					ret[k] = headerWidth
				}
			} else {
				// not header row? column width may not exceed max width
			}
			cellWidth := runeWidth(tbl.rows[i][k])
			if cellWidth > maxColWidth {
				cellWidth = maxColWidth
			}
			if cellWidth > ret[k] {
				ret[k] = cellWidth
			}
		}
	}
	return ret
}

// repeat `s`, `n` times
func repeat(s string, n int) string {
	var ret string
	for i := 0; i < n; i++ {
		ret += s
	}
	return ret
}

// [3,3] -> +---+---+
func stringifyDividingRow(colWidths []int, numLabelLevels int, header bool) string {
	// set dividing symbol values (default: border)
	edge := borderEdge
	labelEdge := borderLabelEdge
	filler := borderFiller
	if header {
		edge = headerEdge
		labelEdge = headerLabelEdge
		filler = headerFiller
	}

	ret := strings.Builder{}
	// leftmost edge
	ret.WriteString(edge)

	for k := range colWidths {
		// sets the number of filler symbols per column, plus a 1-space buffer on either end
		ret.WriteString(repeat(filler, 1+colWidths[k]+1))
		if k == numLabelLevels-1 {
			ret.WriteString(labelEdge)
		} else {
			ret.WriteString(edge)
		}
	}
	return fmt.Sprintln(ret.String())
}

func exceedsMaxWidth(s string, maxWidth int) bool {
	return runeWidth(s) > maxWidth
}

func truncate(s string, maxWidth int) string {
	if !exceedsMaxWidth(s, maxWidth) {
		return s
	}
	r := []rune(s)
	return string(r[:maxWidth-3]) + "..."
}

// try to wrap at a space.
// if wrapping mid-word, insert hyphen
func wrap(s string, maxWidth int) (firstLine string, remainder string) {
	// no split required?
	if !exceedsMaxWidth(s, maxWidth) {
		return s, ""
	}

	r := []rune(s)
	// last letter is whitespace? truncate last whitespace
	if unicode.IsSpace(r[maxWidth-1]) {
		return string(r[:maxWidth-1]), string(r[maxWidth:])
	}
	// penultimate letter is space?
	if unicode.IsSpace(r[maxWidth-2]) {
		// single-character word? retain on line and truncate the next whitespace
		if unicode.IsSpace(r[maxWidth]) {
			return string(r[:maxWidth]), strings.TrimLeftFunc(string(r[maxWidth:]), unicode.IsSpace)
		}
		// truncate last whitesapce
		return string(r[:maxWidth-2]), string(r[maxWidth-1:])
	}
	// multi-character word? insert "-" at end
	ret := make([]rune, maxWidth-1)
	copy(ret, r[:maxWidth-1])
	ret = append(ret, '-')
	return string(ret), string(r[maxWidth-1:])
}

// handle overly-wide columns by either wrapping or truncating.
// if wrapping, writes multiple lines per row.
func (tbl *Table) stringifyContentRow(colWidths []int, content []string, header bool) string {
	// loop until there are no remaining wrapped lines to print
	ret := strings.Builder{}
	for {
		var moreWrappedLines bool

		// leftmost edge
		ret.WriteString(contentEdge)

		// iterate over columns
		for k := range colWidths {
			var remainder string
			// handling overly-wide columns
			if exceedsMaxWidth(content[k], colWidths[k]) {
				// truncate?
				if tbl.truncateCells {
					content[k] = truncate(content[k], colWidths[k])
				} else {
					// wrap?
					var firstLine string
					firstLine, remainder = wrap(content[k], colWidths[k])
					if remainder != "" {
						moreWrappedLines = true
					}
					content[k] = firstLine
				}
			}
			// Center the content in header rows. Use Table alignment (default: Center) for non-header rows.
			alignment := tbl.alignment
			if header && tbl.autoCenterHeaders {
				alignment = AlignCenter
			}
			// align text content and add to string
			ret.WriteString(alignString(content[k], colWidths[k], alignment))
			// add separator after column, including at rightmost edge
			if k == tbl.numLabelLevels-1 {
				ret.WriteString(contentLabelEdge)
			} else {
				ret.WriteString(contentEdge)
			}
			// overwrite content with either wrappedLine or empty cell
			content[k] = remainder
		}
		// start a new line if text is wrapped, otherwise end the loop
		if moreWrappedLines {
			ret.WriteString("\n")
		} else {
			break
		}
	}

	return fmt.Sprintln(ret.String())
}

// expects string to already be truncated or wrapped.
// adds a 1-space buffer on either side
func alignString(s string, width int, alignment Alignment) string {
	if alignment == AlignLeft {
		return fmt.Sprintf(" %-*s ", width, s)
	}
	if alignment == AlignRight {
		return fmt.Sprintf(" %*s ", width, s)
	}
	rightJustified := fmt.Sprintf("%*s", (width+runeWidth(s))/2, s)
	return fmt.Sprintf(" %-*s ", width, rightJustified)
}
