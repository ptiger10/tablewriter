package tablewriter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

// de-couple tests from global variables
func TestMain(m *testing.M) {
	dividingEdge = "+"
	dividingLabelEdge = "++"
	dividingFiller = "-"
	contentEdge = "|"
	labelEdge = "||"

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestTable_render(t *testing.T) {
	type fields struct {
		w              io.Writer
		rows           [][]string
		alignment      Alignment
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"no labels - no header - auto merge",
			fields{
				rows:           [][]string{{"foo", "bar"}, {"foo", "quux"}, {"baz", "quux"}},
				alignment:      AlignLeft,
				numHeaderRows:  0,
				numLabelLevels: 0,
				autoMerge:      true},
			"" +
				"+-----+------+\n" +
				"| foo | bar  |\n" +
				"|     | quux |\n" +
				"| baz |      |\n" +
				"+-----+------+\n",
			false,
		},
		{"no labels - 1 header - no auto merge",
			fields{
				rows:           [][]string{{"foo", "bar"}, {"corge", "quux"}, {"baz", "fred"}},
				alignment:      AlignLeft,
				numHeaderRows:  1,
				numLabelLevels: 0},
			"" +
				"+-------+------+\n" +
				"| foo   | bar  |\n" +
				"+-------+------+\n" +
				"+-------+------+\n" +
				"| corge | quux |\n" +
				"| baz   | fred |\n" +
				"+-------+------+\n",
			false,
		},
		{"labels - no header - no auto merge",
			fields{
				rows:           [][]string{{"foo", "bar"}, {"corge", "quux"}, {"baz", "fred"}},
				alignment:      AlignLeft,
				numHeaderRows:  0,
				numLabelLevels: 1},
			"" +
				"+-------++------+\n" +
				"| foo   || bar  |\n" +
				"| corge || quux |\n" +
				"| baz   || fred |\n" +
				"+-------++------+\n",
			false,
		},
		{"labels & header - no auto merge",
			fields{
				rows:           [][]string{{"foo", "bar"}, {"corge", "quux"}, {"baz", "fred"}},
				alignment:      AlignLeft,
				numHeaderRows:  1,
				numLabelLevels: 1},
			"" +
				"+-------++------+\n" +
				"| foo   || bar  |\n" +
				"+-------++------+\n" +
				"+-------++------+\n" +
				"| corge || quux |\n" +
				"| baz   || fred |\n" +
				"+-------++------+\n",
			false,
		},
		{"fail - no data",
			fields{
				rows:           [][]string{},
				alignment:      AlignLeft,
				numHeaderRows:  0,
				numLabelLevels: 0},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				alignment:      tt.fields.alignment,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			got, err := tbl.render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Table.render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Table.render() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testBadWriter string

func (w testBadWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("err")
}

func TestTable_Render(t *testing.T) {

	type fields struct {
		w              io.Writer
		rows           [][]string
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"bytes.Buffer",
			fields{
				w:              new(bytes.Buffer),
				rows:           [][]string{{"foo", "bar"}},
				numHeaderRows:  0,
				numLabelLevels: 0,
				autoMerge:      false,
				truncateCells:  false},
			false,
		},
		{"fail - bad writer",
			fields{
				w:              testBadWriter(""),
				rows:           [][]string{{"foo", "bar"}},
				numHeaderRows:  0,
				numLabelLevels: 0,
				autoMerge:      false,
				truncateCells:  false},
			true,
		},
		{"fail - empty table",
			fields{
				w:              new(bytes.Buffer),
				rows:           [][]string{},
				numHeaderRows:  0,
				numLabelLevels: 0,
				autoMerge:      false,
				truncateCells:  false},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			err := tbl.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Table.render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestTable_resizeColWidths(t *testing.T) {
	type fields struct {
		w              io.Writer
		rows           [][]string
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []int
	}{
		{"no headers",
			fields{
				rows:          [][]string{{"foo", "baaz", "111111111111111111111111111111111111111111111"}},
				numHeaderRows: 0,
			},
			[]int{3, 4, maxColWidth},
		},
		{"headers",
			fields{
				rows:          [][]string{{"foo", "baaz", "111111111111111111111111111111111111111111111"}},
				numHeaderRows: 1,
			},
			[]int{3, 4, 45},
		},
		{"multiple rows - cell longer than header",
			fields{
				rows:          [][]string{{"foo"}, {"quux"}, {"corge"}},
				numHeaderRows: 1,
			},
			[]int{5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			if got := tbl.resizeColWidths(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Table.resizeColWidths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_alignString(t *testing.T) {
	type args struct {
		s         string
		maxWidth  int
		alignment Alignment
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// NB: there is also a 1-space buffer on either side!
		{name: "left",
			args: args{"foo", 10, AlignLeft},
			want: " foo        ",
		},
		{name: "right",
			args: args{"foo", 10, AlignRight},
			want: "        foo ",
		},
		{name: "center",
			args: args{"foo", 9, AlignCenter},
			want: "    foo    "},
		{name: "center - odd spaces - more to the left",
			args: args{"foo", 6, AlignCenter},
			want: "  foo   ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := alignString(tt.args.s, tt.args.maxWidth, tt.args.alignment); got != tt.want {
				t.Errorf("alignString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_truncate(t *testing.T) {
	type args struct {
		s        string
		maxWidth int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no truncate required", args{"much too long", 13}, "much too long"},
		{"ASCII", args{"much too long indeed", 10}, "much to..."},
		{"non-ASCII", args{"å¬ßø too long", 10}, "å¬ßø to..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncate(tt.args.s, tt.args.maxWidth); got != tt.want {
				t.Errorf("truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wrap(t *testing.T) {
	type args struct {
		s        string
		maxWidth int
	}
	tests := []struct {
		name          string
		args          args
		wantLine      string
		wantRemainder string
	}{
		{"no split", args{"much too long", 13}, "much too long", ""},
		{"split before space", args{"much too long indeed", 9}, "much too", "long indeed"},
		{"split after first letter after a penultimate space, if it is a single-character word ",
			args{"keep the 1 though", 10}, "keep the 1", "though"},
		{"split before first letter after a penultimate space, if it is a multi-character word",
			args{"much too long indeed", 10}, "much too", "long indeed"},
		{"split midword", args{"much too long indeed", 7}, "much t-", "oo long indeed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := wrap(tt.args.s, tt.args.maxWidth)
			if got != tt.wantLine {
				t.Errorf("wrap() = %v, want %v", got, tt.wantLine)
			}
			if got1 != tt.wantRemainder {
				t.Errorf("wrap() remainder = %v, want %v", got1, tt.wantRemainder)
			}
		})
	}
}

func Test_stringifyDividingRow(t *testing.T) {
	type args struct {
		columnWidths   []int
		numLabelLevels int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"no label levels",
			args{[]int{1, 3, 1}, 0},
			"+---+-----+---+\n",
		},
		{
			"1 label level",
			args{[]int{1, 3, 1}, 1},
			"+---++-----+---+\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringifyDividingRow(tt.args.columnWidths, tt.args.numLabelLevels); got != tt.want {
				t.Errorf("stringifyDividingRow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_stringifyContentRow(t *testing.T) {
	type fields struct {
		w              io.Writer
		rows           [][]string
		alignment      Alignment
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	type args struct {
		colWidths []int
		content   []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet string
	}{
		{"no labels - all rows 1 line",
			fields{
				rows:           [][]string{{"foo", "bar"}, {"baz", "qux"}},
				alignment:      AlignCenter,
				numHeaderRows:  0,
				numLabelLevels: 0,
				truncateCells:  false},
			args{
				[]int{3, 3}, []string{"foo", "bar"},
			},
			"| foo | bar |\n",
		},
		{"no labels - wrap & split to newline",
			fields{
				rows:           [][]string{{"foo", "bar"}, {"baz", "qux"}},
				alignment:      AlignCenter,
				numHeaderRows:  0,
				numLabelLevels: 0,
				truncateCells:  false},
			args{
				[]int{3, 2}, []string{"foo", "bar"},
			},
			"" +
				"| foo | b- |\n" +
				"|     | ar |\n",
		},
		{"no labels - truncate",
			fields{
				rows:           [][]string{{"foo", "corge"}, {"baz", "qux"}},
				alignment:      AlignCenter,
				numHeaderRows:  0,
				numLabelLevels: 0,
				truncateCells:  true},
			args{
				[]int{3, 4}, []string{"foo", "corge"},
			},
			"| foo | c... |\n",
		},
		{"1 label level - all rows 1 line",
			fields{
				rows:           [][]string{{"foo", "bar"}, {"baz", "qux"}},
				alignment:      AlignCenter,
				numHeaderRows:  0,
				numLabelLevels: 1,
				autoMerge:      false,
				truncateCells:  false},
			args{
				[]int{3, 3}, []string{"foo", "bar"},
			},
			"| foo || bar |\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				alignment:      tt.fields.alignment,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			if gotRet := tbl.stringifyContentRow(tt.args.colWidths, tt.args.content); gotRet != tt.wantRet {
				t.Errorf("Table.stringifyContentRow() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func Test_autoMergeRows(t *testing.T) {
	type args struct {
		priorRow   []string
		currentRow []string
	}
	tests := []struct {
		name        string
		args        args
		wantPrior   []string
		wantCurrent []string
	}{
		{name: "pass",
			args:        args{[]string{"foo", "bar"}, []string{"baz", "bar"}},
			wantPrior:   []string{"baz", "bar"},
			wantCurrent: []string{"baz", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			autoMergeRows(tt.args.priorRow, tt.args.currentRow)
			if !reflect.DeepEqual(tt.args.priorRow, tt.wantPrior) {
				t.Errorf("autoMergeRows() priorRow -> %v, want %v", tt.args.priorRow, tt.wantPrior)
			}
			if !reflect.DeepEqual(tt.args.currentRow, tt.wantCurrent) {
				t.Errorf("autoMergeRows() currentRow -> %v, want %v", tt.args.currentRow, tt.wantCurrent)
			}
		})
	}
}

func TestNewTable(t *testing.T) {
	tests := []struct {
		name  string
		want  *Table
		wantW string
	}{
		{"Pass",
			&Table{
				w:    &bytes.Buffer{},
				rows: [][]string{},
			},
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if got := NewTable(w); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTable() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("NewTable() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestTable_sameShape(t *testing.T) {
	type fields struct {
		w              io.Writer
		rows           [][]string
		alignment      Alignment
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	type args struct {
		row []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"pass - same length",
			fields{
				rows: [][]string{{"foo"}}},
			args{[]string{"corge"}},
			false},
		{"pass - empty",
			fields{
				rows: [][]string{}},
			args{[]string{"bar"}},
			false},
		{"fail - different lengths",
			fields{
				rows: [][]string{{"foo"}}},
			args{[]string{"bar", "baz"}},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				alignment:      tt.fields.alignment,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			if err := tbl.sameShape(tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Table.sameShape() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTable_AppendHeaderRow(t *testing.T) {
	type fields struct {
		w              io.Writer
		rows           [][]string
		alignment      Alignment
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	type args struct {
		row []string
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantRows             [][]string
		wantNumberHeaderRows int
		wantErr              bool
	}{
		{"pass - empty table",
			fields{
				rows:          [][]string{},
				numHeaderRows: 0,
			},
			args{[]string{"bar"}},
			[][]string{{"bar"}},
			1,
			false},
		{"pass - existing header only",
			fields{
				rows:          [][]string{{"foo"}},
				numHeaderRows: 1,
			},
			args{[]string{"bar"}},
			[][]string{{"foo"}, {"bar"}},
			2,
			false},
		{"pass - existing non-header row only",
			fields{
				rows:          [][]string{{"foo"}},
				numHeaderRows: 0,
			},
			args{[]string{"bar"}},
			[][]string{{"bar"}, {"foo"}},
			1,
			false},
		{"pass - existing header and non-header rows",
			fields{
				rows:          [][]string{{"foo"}, {"baz"}},
				numHeaderRows: 1,
			},
			args{[]string{"bar"}},
			[][]string{{"foo"}, {"bar"}, {"baz"}},
			2,
			false},
		{"fail - wrong shape",
			fields{
				rows:          [][]string{{"foo"}},
				numHeaderRows: 0,
			},
			args{[]string{"corge", "qux"}},
			[][]string{{"foo"}},
			0,
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				alignment:      tt.fields.alignment,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			if err := tbl.AppendHeaderRow(tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Table.AppendHeaderRow() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tbl.rows, tt.wantRows) {
				t.Errorf("Table.AppendHeaderRow().rows -> %v, want %v", tbl.rows, tt.wantRows)
			}

			if tbl.numHeaderRows != tt.wantNumberHeaderRows {
				t.Errorf("Table.AppendHeaderRow().numHeaderRows -> %v, want %v", tbl.numHeaderRows, tt.fields.numHeaderRows)
			}
		})
	}
}

func TestTable_AppendRow(t *testing.T) {
	type fields struct {
		w              io.Writer
		rows           [][]string
		alignment      Alignment
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	type args struct {
		row []string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantRows [][]string
		wantErr  bool
	}{
		{"pass",
			fields{
				rows: [][]string{{"foo"}},
			},
			args{[]string{"bar"}},
			[][]string{{"foo"}, {"bar"}},
			false},
		{"fail - wrong shape",
			fields{
				rows: [][]string{{"foo"}},
			},
			args{[]string{"corge", "qux"}},
			[][]string{{"foo"}},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				alignment:      tt.fields.alignment,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			if err := tbl.AppendRow(tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Table.AppendRow() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tbl.rows, tt.wantRows) {
				t.Errorf("Table.AppendHeaderRow().rows -> %v, want %v", tbl.rows, tt.wantRows)
			}
		})
	}
}

func TestTable_AppendRows(t *testing.T) {
	type fields struct {
		w              io.Writer
		rows           [][]string
		alignment      Alignment
		separateLabels bool
		numHeaderRows  int
		numLabelLevels int
		autoMerge      bool
		truncateCells  bool
	}
	type args struct {
		rows [][]string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantRows [][]string
		wantErr  bool
	}{
		{"pass",
			fields{
				rows: [][]string{{"foo"}},
			},
			args{[][]string{{"bar"}, {"baz"}}},
			[][]string{{"foo"}, {"bar"}, {"baz"}},
			false},
		{"fail - bad shape",
			fields{
				rows: [][]string{{"foo"}},
			},
			args{[][]string{{"corge", "qux"}}},
			[][]string{{"foo"}},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				w:              tt.fields.w,
				rows:           tt.fields.rows,
				alignment:      tt.fields.alignment,
				separateLabels: tt.fields.separateLabels,
				numHeaderRows:  tt.fields.numHeaderRows,
				numLabelLevels: tt.fields.numLabelLevels,
				autoMerge:      tt.fields.autoMerge,
				truncateCells:  tt.fields.truncateCells,
			}
			if err := tbl.AppendRows(tt.args.rows); (err != nil) != tt.wantErr {
				t.Errorf("Table.AppendRows() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tbl.rows, tt.wantRows) {
				t.Errorf("Table.AppendHeaderRow().rows -> %v, want %v", tbl.rows, tt.wantRows)
			}
		})
	}
}

func Test_SetSymbols(t *testing.T) {
	SetDividingEdge("*")
	if dividingEdge != "*" {
		t.Errorf("SetDividingEdge() failed")
	}

	SetDividingLabelEdge("**")
	if dividingLabelEdge != "**" {
		t.Errorf("SetDividingLabelEdge() failed")
	}

	SetDividingFiller("=")
	if dividingFiller != "=" {
		t.Errorf("SetDividingFiller() failed")
	}

	SetContentEdge("%")
	if contentEdge != "%" {
		t.Errorf("SetContentEdge() failed")
	}

	SetLabelEdge("%%")
	if labelEdge != "%%" {
		t.Errorf("SetLabelEdge() failed")
	}

}

func TestTable_MergeRepeats(t *testing.T) {
	type fields struct {
		autoMerge bool
	}
	tests := []struct {
		name          string
		fields        fields
		wantAutoMerge bool
	}{
		{"pass", fields{autoMerge: false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				autoMerge: tt.fields.autoMerge,
			}
			tbl.MergeRepeats()

			if tbl.autoMerge != tt.wantAutoMerge {
				t.Errorf("Table.MergeRepeats().autoMerge -> %v, want %v", tbl.autoMerge, tt.wantAutoMerge)
			}
		})
	}
}

func TestTable_TruncateCells(t *testing.T) {
	type fields struct {
		autoMerge bool
	}
	tests := []struct {
		name         string
		fields       fields
		wantTruncate bool
	}{
		{"pass", fields{autoMerge: false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := &Table{
				autoMerge: tt.fields.autoMerge,
			}
			tbl.TruncateWideCells()

			if tbl.truncateCells != tt.wantTruncate {
				t.Errorf("Table.TruncateWideCells().truncateCells -> %v, want %v", tbl.truncateCells, tt.wantTruncate)
			}
		})
	}
}
