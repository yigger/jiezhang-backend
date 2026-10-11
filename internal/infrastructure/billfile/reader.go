package billfile

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"github.com/yigger/jiezhang-backend/internal/service/billimport"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type Reader struct{}

func (Reader) Read(ctx context.Context, reader io.Reader, format string) ([]billimport.Table, error) {
	data, err := io.ReadAll(io.LimitReader(reader, billimport.MaxFileBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > billimport.MaxFileBytes {
		return nil, errors.New("bill file too large")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch strings.ToLower(strings.TrimPrefix(format, ".")) {
	case "csv":
		if !utf8.Valid(data) {
			data, err = simplifiedchinese.GB18030.NewDecoder().Bytes(data)
			if err != nil || bytes.Contains(data, []byte("\ufffd")) {
				return nil, errors.New("invalid CSV encoding")
			}
		}
		r := csv.NewReader(bytes.NewReader(data))
		r.FieldsPerRecord = -1
		var table billimport.Table
		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			values, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			line, _ := r.FieldPos(0)
			for len(table.Rows) < line-1 {
				table.Rows = append(table.Rows, nil)
			}
			cells := make([]billimport.Cell, len(values))
			for i, v := range values {
				cells[i].Text = v
			}
			table.Rows = append(table.Rows, cells)
			if len(table.Rows) > billimport.MaxRows+100 {
				return nil, errors.New("too many CSV rows")
			}
		}
		return []billimport.Table{table}, nil
	case "xlsx":
		f, err := excelize.OpenReader(bytes.NewReader(data), excelize.Options{UnzipSizeLimit: 100 << 20, UnzipXMLSizeLimit: 20 << 20})
		if err != nil {
			return nil, err
		}
		defer f.Close()
		props, err := f.GetWorkbookProps()
		if err != nil {
			return nil, err
		}
		var tables []billimport.Table
		var totalRows int
		for _, sheet := range f.GetSheetList() {
			rows, err := f.Rows(sheet)
			if err != nil {
				return nil, err
			}
			table := billimport.Table{Name: sheet, Excel1904: props.Date1904 != nil && *props.Date1904}
			for rows.Next() {
				if err := ctx.Err(); err != nil {
					_ = rows.Close()
					return nil, err
				}
				values, err := rows.Columns(excelize.Options{RawCellValue: true})
				if err != nil {
					_ = rows.Close()
					return nil, err
				}
				cells := make([]billimport.Cell, len(values))
				for i, v := range values {
					ref, _ := excelize.CoordinatesToCellName(i+1, len(table.Rows)+1)
					kind, err := f.GetCellType(sheet, ref)
					if err != nil {
						_ = rows.Close()
						return nil, err
					}
					formula, err := f.GetCellFormula(sheet, ref)
					if err != nil {
						_ = rows.Close()
						return nil, err
					}
					cells[i] = billimport.Cell{Text: v, Numeric: kind == excelize.CellTypeNumber || (kind == excelize.CellTypeUnset && v != ""), Formula: formula != ""}
				}
				table.Rows = append(table.Rows, cells)
				totalRows++
				if totalRows > billimport.MaxRows+100 {
					_ = rows.Close()
					return nil, errors.New("too many XLSX rows")
				}
			}
			if err := rows.Error(); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if err := rows.Close(); err != nil {
				return nil, err
			}
			tables = append(tables, table)
		}
		return tables, nil
	default:
		return nil, errors.New("unsupported bill format")
	}
}
