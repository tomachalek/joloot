package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/czcorpus/cnc-gokit/collections"
)

const (
	ChunkTypeDateTime ChunkType = iota
	ChunkTypeKey
	ChunkTypeTextValue
)

type ChunkType int

type JSONChunk struct {
	Type  ChunkType
	Value any
}

func (jc JSONChunk) IsKey() bool {
	return jc.Type == ChunkTypeKey
}

func (jc JSONChunk) IsDatetime() bool {
	return jc.Type == ChunkTypeDateTime
}

func (jc JSONChunk) IsNil() bool {
	return jc.Value == nil
}

func (jc JSONChunk) StringValue() string {
	switch tv := jc.Value.(type) {
	case string:
		return tv
	case float64:
		if float64(int(tv)) != tv {
			return fmt.Sprintf("%01.2f", tv) // TODO configurable precision
		}
		return fmt.Sprintf("%d", int(tv))
	case bool:
		return fmt.Sprintf("%t", tv)
	case nil:
		return "-"
	default:
		return fmt.Sprintf("#TODO# %v", tv)
	}
}

func (jc JSONChunk) IsMapValue() bool {
	_, ok := jc.Value.(map[string]any)
	return ok
}

type JSONData map[string]any

func (jd JSONData) Chunks(dtField string) []JSONChunk {
	ans := make([]JSONChunk, 0, 2*len(jd))
	return jd.chunksRecursive(dtField, jd, "", ans)
}

func (jd JSONData) chunksRecursive(
	dtField string,
	curr JSONData,
	parentKey string,
	chunks []JSONChunk,
) []JSONChunk {
	keys := make([]string, 0, len(curr))
	var hasDtKey bool
	for k := range curr {
		if k != dtField {
			keys = append(keys, k)

		} else {
			hasDtKey = true
		}
	}
	sort.Strings(keys)
	if hasDtKey {
		keys = append([]string{dtField}, keys...)
	}
	for _, k := range keys {
		if k == dtField {
			chunks = append(
				chunks,
				JSONChunk{Type: ChunkTypeDateTime, Value: curr[k]},
			)
			continue
		}
		v := curr[k]
		switch tv := v.(type) {
		case map[string]any:
			jd.chunksRecursive(dtField, tv, k, chunks[:])
		default:
			key := k
			if parentKey != "" {
				key = parentKey + "." + k
			}
			chunks = append(
				chunks,
				JSONChunk{Type: ChunkTypeKey, Value: key},
				JSONChunk{Type: ChunkTypeTextValue, Value: v},
			)
		}
	}
	return chunks
}

type Line struct {
	Data JSONData
}

type FileNavigator struct {
	file    *os.File
	scanner *bufio.Scanner
	pos     int64
	buffer  *collections.CircularList[*Line]
}

func (fn *FileNavigator) Init(numLines int) error {
	for i := 0; i < numLines; i++ {
		err := fn.NextLine()
		if err != nil {
			return err
		}
	}
	return nil
}

func (fn *FileNavigator) ForItemsBuffer(applyFn func(i int, line *Line) bool) {
	fn.buffer.ForEach(applyFn)
}

func (fn *FileNavigator) NextLine() error {
	fn.pos, _ = fn.file.Seek(0, io.SeekCurrent)
	if fn.scanner.Err() != nil {
		return fn.scanner.Err()

	} else if fn.scanner.Scan() {
		data := make(map[string]any)
		json.Unmarshal(fn.scanner.Bytes(), &data) // TODO handle error
		fn.buffer.Append(&Line{data})
	}
	return nil
}

func (fn *FileNavigator) PreviousLine() error {
	_, err := fn.file.Seek(fn.pos, io.SeekStart) // Go back to saved position
	if err != nil {
		return err
	}
	fn.scanner = bufio.NewScanner(fn.file)      // Create a new scanner
	fn.pos, _ = fn.file.Seek(0, io.SeekCurrent) // Save current position again
	if fn.scanner.Err() != nil {
		return fn.scanner.Err()

	} else if fn.scanner.Scan() {
		data := make(map[string]any)
		json.Unmarshal(fn.scanner.Bytes(), &data) // TODO handle error
		fn.buffer.Prepend(&Line{data})
	}
	return nil
}

func (fn *FileNavigator) Close() {
	fn.file.Close()
}

func NewFileNavigator(path string, numLines int) (*FileNavigator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	return &FileNavigator{
		file:    file,
		scanner: scanner,
		buffer:  collections.NewCircularList[*Line](numLines),
	}, nil
}
