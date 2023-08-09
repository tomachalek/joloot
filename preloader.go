package main

import (
	"bufio"
	"errors"
	"os"
)

var (
	ErrNoMoreLines = errors.New("no more lines available")
)

type JFile struct {
	file                *os.File
	scanner             *bufio.Scanner
	linesPerPage        int
	currLineNum         int
	currAbsoluteLineNum int
	prevPage            [][]byte
	currPage            [][]byte
	nextPage            [][]byte
}

func (f *JFile) NextLine() ([]byte, error) {
	if f.currLineNum+1 >= len(f.currPage) {
		if len(f.nextPage) == 0 {
			return []byte{}, ErrNoMoreLines
		}
		f.prevPage = f.currPage
		f.currPage = f.nextPage
		f.currAbsoluteLineNum++
		f.currLineNum = 0
		f.nextPage = f.loadPage(f.currAbsoluteLineNum + 1)

		return f.currPage[f.currLineNum], nil

	} else {
		f.currAbsoluteLineNum++
		f.currLineNum++
		return f.currPage[f.currLineNum], nil
	}
}

func (f *JFile) PrevLine() ([]byte, error) {
	if f.currLineNum-1 < 0 {
		if len(f.prevPage) == 0 {
			return []byte{}, ErrNoMoreLines
		}
		f.nextPage = f.currPage
		f.currPage = f.prevPage
		f.prevPage = f.loadPage(f.currAbsoluteLineNum - 1)
		f.currLineNum = f.linesPerPage - 1

	} else {
		f.currLineNum--
	}
	f.currAbsoluteLineNum--
	return f.currPage[f.currLineNum], nil
}

func (f *JFile) loadPage(absLine int) [][]byte {
	return [][]byte{} // TODO
}

func OpenFile(path string, linesPerPage int) (*JFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	return &JFile{
		file:         file,
		scanner:      scanner,
		linesPerPage: linesPerPage,
	}, nil
}
