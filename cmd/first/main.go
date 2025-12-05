package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	MaxFileSizeAllowed int64 = 1_000_000
	DefaultStart       int   = 50
	MaxDial            int   = 99
	MinDial            int   = 0
	TargetDial         int   = 0
)

type Parser interface {
	Parse() error
}

type FileReader struct {
	filePath string
	reader   *bufio.Reader
}

func (r *FileReader) Read() (string, error) {
	if r.filePath == "" {
		return "", fmt.Errorf("File path is empty")
	}
	file, err := os.Open(r.filePath)
	if err != nil {
		return "", fmt.Errorf("Error opening file: %v", err)
	}
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("Error getting file info: %v", err)
	}
	name, size := info.Name(), info.Size()
	if size == 0 {
		return "", fmt.Errorf("%s is empty", name)
	}
	if size > MaxFileSizeAllowed {
		return "", fmt.Errorf("%s is too large", name)
	}
	defer file.Close()
	r.reader = bufio.NewReader(file)
	buffer := []byte{}
	_, error := r.reader.Read(buffer)
	if error != nil {
		return "", fmt.Errorf("Error reading file: %v", err)
	}
	return string(buffer), nil
}

type FirstEventFileParser struct {
	reader        FileReader
	contents      []string
	parsedContent []DialInstruction
}

func (f FirstEventFileParser) readFileToBuffer() error {
	buffer := []byte{}
	_, error := f.reader.Read()
	if error != nil {
		return fmt.Errorf("Error reading file: %v", error)
	}
	f.contents = strings.Split(string(buffer), "\n")
	return nil
}

func (f FirstEventFileParser) parseContent(content string) error {
	if len(content) == 0 {
		return fmt.Errorf("content is empty")
	}
	parseRotation := func(content string) (RotationType, int) {
		rotationTypeString, rotationAmountString := content[0], content[1:]
		if string(rotationTypeString) == "" || rotationAmountString == "" {
			return "", 0
		}
		rotationAmount, err := strconv.Atoi(rotationAmountString)
		if err != nil {
			return "", 0
		}
		return RotationType(rotationTypeString), rotationAmount
	}
	rotationType, rotationAmount := parseRotation(content)
	if rotationType == "" || rotationAmount == 0 {
		return fmt.Errorf("invalid rotation")
	}
	f.parsedContent = append(f.parsedContent, DialInstruction{rotationAmount, rotationType})
	return nil
}

func (f FirstEventFileParser) Parse() error {
	if err := f.readFileToBuffer(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	for idx, content := range f.contents {
		if err := f.parseContent(content); err != nil {
			fmt.Printf("error parsing content at index %d: %w\n", idx, err)
		}
	}
	return nil
}

type RotationType string

const (
	LeftRotation  RotationType = "L"
	RightRotation RotationType = "R"
)

func (r RotationType) GetDialByRotationType(currentDial int) int {
	switch r {
	case LeftRotation:
		return currentDial * -1
	case RightRotation:
		return currentDial * 1
	}
	err := fmt.Sprintf("Invalid rotation type: %s\n", r)
	log.Fatal(err)
	return 0
}

type DialInstruction struct {
	dialRotationAmount int
	typeOfRotation     RotationType
}

type originalString string
type DialInstructions struct {
	Instructions map[originalString]DialInstruction
}
