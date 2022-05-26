package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
)

// go:embed VERSION
var version string

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "taste FILE",
		Version: version,
		Short:   "Print one random row from a Parquet or CSV file as JSON.",
		Long: strings.Join([]string{
			"Print one random row from a Parquet or CSV file as JSON.",
			"",
			"If the input file is a Parquet file, only return row from the first row group.",
			"If the input file is a CSV file, the entire file will be read into memory.",
		}, "\n"),
		Example: strings.Join([]string{
			"# print one random row from a parquet file",
			"taste data.parquet",
			"",
			"# print one random row from a CSV file",
			"taste data.csv",
			"",
			"# use custom CSV delimiter",
			"taste data.csv -d '|'",
			"",
			"# pretty print JSON",
			"taste data.parquet --pretty",
			"",
			"# read 20 custom rows",
			"taste data.parquet -n 20",
		}, "\n"),
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{cobra.BashCompFilenameExt},
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := args[0]
			pretty, err := cmd.Flags().GetBool("pretty")
			if err != nil {
				return err
			}
			delimiter, err := cmd.Flags().GetString("delimiter")
			if err != nil {
				return err
			}
			num, err := cmd.Flags().GetInt("num-rows")
			if err != nil {
				return err
			}
			parts := strings.Split(fp, ".")
			ext := parts[len(parts)-1]
			var rows interface{}
			switch ext {
			case "parquet", "parq":
				rows, err = parquetRow(fp, num)
			case "csv":
				rows, err = csvRow(fp, num, delimiter)
			default:
				return fmt.Errorf("unrecognized extension for file %q", fp)
			}
			if err != nil {
				return err
			}

			var b []byte
			if pretty {
				b, err = json.MarshalIndent(rows, "", "    ")
				if err != nil {
					return fmt.Errorf("json.MarshalIndent err: %v", err)
				}
			} else {
				b, err = json.Marshal(rows)
				if err != nil {
					return fmt.Errorf("json.Marshal err: %v", err)
				}
			}
			cmd.Println(string(b))
			return nil
		},
	}
	cmd.Flags().BoolP("pretty", "p", false, "pretty-print JSON")
	cmd.Flags().StringP("delimiter", "d", "", "CSV delimiter")
	cmd.Flags().IntP("num-rows", "n", 1, "number of rows to return")
	return cmd
}

func rowObjAsMap(pr *reader.ParquetReader, obj interface{}) map[string]interface{} {
	v := reflect.ValueOf(obj)
	m := map[string]interface{}{}
	t := v.Type()
	n := t.NumField()
	for i := 0; i < n; i++ {
		m[pr.SchemaHandler.GetExName(i+1)] = v.FieldByIndex([]int{i}).Interface()
	}
	return m
}

func indices(total int64, n int) (sl []int, err error) {
	sl = make([]int, n)
	for i := 0; i < n; i++ {
		m, err := rand.Int(rand.Reader, big.NewInt(total))
		if err != nil {
			return nil, fmt.Errorf("rand.Int err: %v", err)
		}
		sl[i] = int(m.Int64())
	}
	sort.Slice(sl, func(i, j int) bool {
		return sl[i] < sl[j]
	})
	return sl, nil
}

func parquetRow(fp string, num int) (interface{}, error) {
	pf, err := local.NewLocalFileReader(fp)
	if err != nil {
		return nil, fmt.Errorf("read parquet file err: %v", err)
	}
	defer pf.Close()
	pr, err := reader.NewParquetReader(pf, nil, 4)
	if err != nil {
		return nil, fmt.Errorf("create parquet reader err: %v", err)
	}
	defer pr.ReadStop()

	ind, err := indices(pr.Footer.RowGroups[0].NumRows, num)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, num)
	var last int
	for i, j := range ind {
		if err = pr.SkipRows(int64(j - last)); err != nil {
			return nil, fmt.Errorf("SkipRows err: %v", err)
		}
		sl, err := pr.ReadByNumber(1)
		if err != nil {
			return nil, fmt.Errorf("ReadByNumber err: %v", err)
		}
		rows[i] = rowObjAsMap(pr, sl[0])
		last = j + 1
	}
	if len(rows) > 1 {
		return rows, nil
	}
	return rows[0], nil
}

func csvRow(fp string, num int, delimiter string) (interface{}, error) {
	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("os.Open err: %v", err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	if delimiter != "" {
		r.Comma, _ = utf8.DecodeRune([]byte(delimiter))
	}
	allRows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadAll err: %v", err)
	}
	columns := allRows[0]
	allRows = allRows[1:]
	ind, err := indices(int64(len(allRows)), num)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]string, num)
	for i, j := range ind {
		m := map[string]string{}
		for k, v := range allRows[j] {
			m[columns[k]] = v
		}
		rows[i] = m
	}
	if len(rows) > 1 {
		return rows, nil
	}
	return rows[0], nil
}
