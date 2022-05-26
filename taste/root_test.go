package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pckhoi/taste/taste/random"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

func testCmd(args ...string) *cobra.Command {
	cmd := rootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd
}

func parseCmdJSON(t *testing.T, cmd *cobra.Command, obj interface{}) {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	cmd.SetOut(buf)
	require.NoError(t, cmd.Execute())
	require.NoError(t, json.Unmarshal(buf.Bytes(), obj), buf.String())
}

type Student struct {
	Name   string  `json:"name" parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Age    int32   `json:"age" parquet:"name=age, type=INT32"`
	Id     int64   `json:"id" parquet:"name=id, type=INT64"`
	Weight float32 `json:"weight" parquet:"name=weight, type=FLOAT"`
	Sex    bool    `json:"sex" parquet:"name=sex, type=BOOLEAN"`
}

var csvColumns = []string{
	"name", "age", "id", "weight", "sex",
}

func randomStudents(n int) []Student {
	stus := make([]Student, n)
	for i := 0; i < n; i++ {
		stus[i] = Student{
			Name:   random.AlphaString(10),
			Age:    random.Int32(),
			Id:     int64(i),
			Weight: random.Float32(),
			Sex:    random.Bool(),
		}
	}
	return stus
}

func studentsMap(sl []Student) map[int64]Student {
	m := map[int64]Student{}
	for _, st := range sl {
		m[st.Id] = st
	}
	return m
}

func writeParquet(t *testing.T, fp string, stus []Student) {
	t.Helper()
	fw, err := local.NewLocalFileWriter(fp)
	require.NoError(t, err)
	defer fw.Close()
	pw, err := writer.NewParquetWriter(fw, new(Student), 1)
	require.NoError(t, err)
	defer pw.WriteStop()
	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.CompressionType = parquet.CompressionCodec_SNAPPY
	for _, stu := range stus {
		require.NoError(t, pw.Write(stu))
	}
}

func studentToMap(stu Student) map[string]string {
	return map[string]string{
		"name":   stu.Name,
		"age":    fmt.Sprintf("%d", stu.Age),
		"id":     fmt.Sprintf("%d", stu.Id),
		"weight": fmt.Sprintf("%f", stu.Weight),
		"sex":    fmt.Sprintf("%v", stu.Sex),
	}
}

func writeCSV(t *testing.T, fp string, stus []Student, comma rune) {
	t.Helper()
	f, err := os.Create(fp)
	require.NoError(t, err)
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if comma != 0 {
		w.Comma = comma
	}
	require.NoError(t, w.Write(csvColumns))
	for _, stu := range stus {
		m := studentToMap(stu)
		row := make([]string, len(csvColumns))
		for i, k := range csvColumns {
			row[i] = m[k]
		}
		require.NoError(t, w.Write(row))
	}
}

func TestUnrecognizedFileType(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.txt")
	require.NoError(t, ioutil.WriteFile(fp, []byte("abcdeef"), 0644))

	cmd := testCmd(fp)
	assert.Error(t, cmd.Execute())
}

func TestReadParquet(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.parquet")
	stus := randomStudents(20)
	stusMap := studentsMap(stus)
	writeParquet(t, fp, stus)

	cmd := testCmd(fp)
	stu := &Student{}
	parseCmdJSON(t, cmd, stu)
	require.NotEmpty(t, stu)
	assert.Contains(t, stusMap, stu.Id)
	assert.Equal(t, stusMap[stu.Id], *stu)
}

func TestReadCSV(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.csv")
	stus := randomStudents(20)
	stusMap := studentsMap(stus)
	writeCSV(t, fp, stus, 0)

	cmd := testCmd(fp)
	stu := map[string]string{}
	parseCmdJSON(t, cmd, &stu)
	require.NotEmpty(t, stu)
	id, err := strconv.ParseInt(stu["id"], 10, 64)
	require.NoError(t, err)
	assert.Contains(t, stusMap, id)
	assert.Equal(t, studentToMap(stusMap[id]), stu)
}

func TestReadBatch(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.parquet")
	stus := randomStudents(100)
	stusMap := studentsMap(stus)
	writeParquet(t, fp, stus)

	cmd := testCmd(fp, "-n", "10")
	sl := []Student{}
	parseCmdJSON(t, cmd, &sl)
	require.NotEmpty(t, sl)
	assert.Len(t, sl, 10)
	for _, stu := range sl {
		assert.Contains(t, stusMap, stu.Id)
		assert.Equal(t, stusMap[stu.Id], stu)
	}
}

func TestCSVDelimiter(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.csv")
	stus := randomStudents(20)
	stusMap := studentsMap(stus)
	writeCSV(t, fp, stus, '|')

	cmd := testCmd(fp, "--delimiter", "|")
	stu := map[string]string{}
	parseCmdJSON(t, cmd, &stu)
	require.NotEmpty(t, stu)
	id, err := strconv.ParseInt(stu["id"], 10, 64)
	require.NoError(t, err)
	assert.Contains(t, stusMap, id)
	assert.Equal(t, studentToMap(stusMap[id]), stu)
}
