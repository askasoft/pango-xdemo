package csvutil

import (
	"unicode"

	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
)

type CsvColumn struct {
	Txt string
	Col *int
}

type CsvHeader struct {
	Locales []string
	Columns []*CsvColumn
	Others  map[string]int
}

func (ch *CsvHeader) AddColumn(txt string, col *int) {
	ch.Columns = append(ch.Columns, &CsvColumn{txt, col})
}

func (ch *CsvHeader) findColumn(i int, s string) bool {
	for _, loc := range ch.Locales {
		for _, c := range ch.Columns {
			if str.EqualFold(s, tbs.GetText(loc, c.Txt)) {
				*c.Col = i
				return true
			}
		}
	}
	return false
}

func (ch *CsvHeader) ParseHead(row []string) {
	for _, c := range ch.Columns {
		*c.Col = -1
	}
	ch.Others = make(map[string]int)

	for i, s := range row {
		s = str.Strip(s)
		if s == "" {
			continue
		}

		if !ch.findColumn(i, s) {
			ch.Others[s] = i
		}
	}
}

func GetColumn(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return str.Strip(row[idx])
	}
	return ""
}

func GetString(row []string, idx int) string {
	return str.RemoveFunc(GetColumn(row, idx), func(r rune) bool {
		return r < ' '
	})
}

func GetStrings(row []string, idx int) []string {
	ss := str.FieldsAny(GetColumn(row, idx), "\r\n")
	for i := 0; i < len(ss); i++ {
		ss[i] = str.Strip(ss[i])
	}
	return str.RemoveEmpties(ss)
}

func GetTags(row []string, idx int) []string {
	return str.FieldsFunc(GetColumn(row, idx), func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
}
