package main

import (
	"errors"
	"flag"
	"fmt"
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
}

func (r *FileReader) Read() (string, error) {
	if r.filePath == "" {
		return "", fmt.Errorf("File path is empty")
	}
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return "", fmt.Errorf("Error reading file: %v", err)
	}
	return string(data), nil
}

type FirstEventFileParser struct {
	reader         FileReader
	contents       []string
	parsedContents []DialInstruction
}

func (f *FirstEventFileParser) readFileToBuffer() error {
	content, error := f.reader.Read()
	if error != nil {
		return fmt.Errorf("Error reading file: %v", error)
	}
	f.contents = strings.Split(content, "\n")
	return nil
}

func (f *FirstEventFileParser) parseContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return fmt.Errorf("content is empty")
	}
	parseRotation := func(content string) (RotationType, int) {
		rotationTypeString, rotationAmountString := string(content[0]), content[1:]
		if len(rotationTypeString) == 0 || len(rotationAmountString) == 0 {
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
	f.parsedContents = append(f.parsedContents, DialInstruction{rotationAmount, rotationType})
	return nil
}

func (f *FirstEventFileParser) Parse() error {
	if err := f.readFileToBuffer(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	for idx, content := range f.contents {
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			continue
		}
		if err := f.parseContent(trimmed); err != nil {
			fmt.Printf("error parsing content at index %d: %v\n", idx, err)
		}
	}
	return nil
}

type RotationType string

const (
	LeftRotation  RotationType = "L"
	RightRotation RotationType = "R"
)

func (r RotationType) Apply(amount int) int {
	switch r {
	case LeftRotation:
		return amount * -1
	case RightRotation:
		return amount
	}
	fmt.Printf("Invalid rotation type: %s\n", r)
	return 0
}

type DialInstruction struct {
	dialRotationAmount int
	typeOfRotation     RotationType
}

type originalString string
type DialInstructions struct {
	parser       FirstEventFileParser
	Instructions []DialInstruction
}

func NewDialInstructions(parser FirstEventFileParser) DialInstructions {
	return DialInstructions{
		parser:       parser,
		Instructions: make([]DialInstruction, 0),
	}
}

func (d *DialInstructions) Seed() error {
	if err := d.parser.Parse(); err != nil {
		return err
	}
	for _, content := range d.parser.parsedContents {
		d.Instructions = append(d.Instructions, DialInstruction{
			dialRotationAmount: content.dialRotationAmount,
			typeOfRotation:     content.typeOfRotation,
		})
	}
	return nil
}

func (d DialInstructions) FindPasswords() (int, error) {
	if len(d.Instructions) == 0 {
		return 0, errors.New("no instructions found")
	}
	var foundPasswordCounter int
	currentDial := DefaultStart
	dialRange := MaxDial + 1
	for _, instruction := range d.Instructions {
		rotation := instruction.typeOfRotation.Apply(instruction.dialRotationAmount)
		currentDial = (currentDial + rotation) % dialRange
		if currentDial < 0 {
			currentDial += dialRange
		}
		if currentDial == TargetDial {
			foundPasswordCounter += 1
		}
	}
	return foundPasswordCounter, nil
}

func main() {
	instructionsPath := flag.String("instructions", "cmd/first/instructions.txt", "path to the instruction file")
	flag.Parse()

	reader := FileReader{filePath: *instructionsPath}
	parser := FirstEventFileParser{reader: reader}
	instructions := NewDialInstructions(parser)
	if err := instructions.Seed(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to seed instructions: %v\n", err)
		os.Exit(1)
	}
	count, err := instructions.FindPasswords()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find password: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Password appears %d times\n", count)
}
