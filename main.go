package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"
)

const progStart = 0x200

func isValidReg(regStr string) bool {
	re, _ := regexp.Compile(`v([1-9]|1[0-6])`)
	return re.Match([]byte(regStr))
}

func getRegNum(regStr string) uint16 {
	split := strings.Split(regStr, "v")
	regNum, _ := strconv.ParseUint(split[1], 10, 16)
	return uint16(regNum)
}

func parseReg(s *scanner.Scanner) uint16 {
	s.Scan()
	regStr := s.TokenText()
	isReg := isValidReg(regStr)
	if !isReg {
		errStr := fmt.Errorf("invalid register operand %s at line %d", regStr, s.Pos().Line)
		panic(errStr)
	}

	return getRegNum(regStr)
}

func main() {

	var inFile string
	var outFile string
	flag.StringVar(&inFile, "i", "", "The Chip8 asm file to assemble.")
	flag.StringVar(&outFile, "o", "", "The assembled Chip8 binary.")
	flag.Parse()

	if inFile == "" {
		fmt.Fprintf(os.Stderr, "Error: No input file provided!\n")
		os.Exit(1)

	}
	if outFile == "" {
		fmt.Fprintf(os.Stderr, "Error: No output filename specified!\n")
		os.Exit(1)
	}

	rom := make([]uint16, 0)
	symbolTable := make(map[string]uint16)
	undefTable := make(map[string][]int)
	src, err := ioutil.ReadFile(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
	var s scanner.Scanner
	s.Init(bytes.NewBuffer(src))
	labeledLine := false
	currInstNum := 0
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		// fmt.Printf("pos: %s text: %s\n", s.Position, s.TokenText())
		text := strings.ToLower(s.TokenText())
		switch text {
		case "rts":
			rom = append(rom, 0x00EE)
			labeledLine = false
		case "clr":
			rom = append(rom, 0x00E0)
			labeledLine = false
		case "jump":
			tok = s.Scan()
			jumpAddr := symbolTable[s.TokenText()]
			if jumpAddr == 0 {
				undefTable[s.TokenText()] = append(undefTable[s.TokenText()], currInstNum)
				rom = append(rom, 0x1000)
			} else {
				rom = append(rom, 0x1000|symbolTable[s.TokenText()])
				labeledLine = false
			}
			labeledLine = false
		case "call":
			tok = s.Scan()
			rom = append(rom, 0x2000|symbolTable[s.TokenText()])
			labeledLine = false
		case "ske":
			tok = s.Scan()
			regStr := s.TokenText()
			isReg := isValidReg(regStr)
			if !isReg {
				errStr := fmt.Errorf("invalid register operand %s at line %d", regStr, s.Pos().Line)
				panic(errStr)
			}

			split := strings.Split(regStr, "v")
			regNum, _ := strconv.ParseUint(split[1], 10, 16)
			regNum16 := uint16(regNum)
			// fmt.Printf("reg: %#x\n", regNum16)
			// now get the second operand
			tok = s.Scan()
			imm64, _ := strconv.ParseUint(s.TokenText(), 10, 16)
			imm := uint16(imm64)

			rom = append(rom, 0x3000|regNum16<<8|imm)
			labeledLine = false

		case "skne":
			tok = s.Scan()
			regStr := s.TokenText()
			isReg := isValidReg(regStr)
			if !isReg {
				errStr := fmt.Errorf("invalid register operand %s at line %d", regStr, s.Pos().Line)
				panic(errStr)
			}

			split := strings.Split(regStr, "v")
			regNum, _ := strconv.ParseUint(split[1], 10, 16)
			regNum16 := uint16(regNum)
			// fmt.Printf("reg: %#x\n", regNum16)
			// now get the second operand
			tok = s.Scan()
			imm64, _ := strconv.ParseUint(s.TokenText(), 10, 16)
			imm := uint16(imm64)

			rom = append(rom, 0x4000|regNum16<<8|imm)
			labeledLine = false

		case "skre":
			tok = s.Scan()
			reg1Str := s.TokenText()

			if !isValidReg(reg1Str) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", reg1Str, s.Pos().Line)
				panic(errStr)
			}
			tok = s.Scan()
			reg2Str := s.TokenText()

			if !isValidReg(reg2Str) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", reg2Str, s.Pos().Line)
				panic(errStr)
			}

			reg1Num := getRegNum(reg1Str)
			reg2Num := getRegNum(reg2Str)

			rom = append(rom, 0x5000|reg1Num<<8|reg2Num<<4)
			labeledLine = false
		case "load":
			tok = s.Scan()
			srcReg := s.TokenText()
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			immStr := s.TokenText()
			var imm uint64
			if immStr[0] == '0' && immStr[1] == 'x' {
				imm, _ = strconv.ParseUint(immStr[2:], 16, 16)
			} else {

				imm, _ = strconv.ParseUint(immStr, 10, 16)
			}
			imm16 := uint16(imm)

			rom = append(rom, 0x6000|srcRegNum<<8|imm16)
			labeledLine = false
		case "add":
			tok = s.Scan()
			srcReg := s.TokenText()
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			immStr := s.TokenText()
			var imm uint64
			if immStr[0] == '0' && immStr[1] == 'x' {
				imm, _ = strconv.ParseUint(immStr[2:], 16, 16)
			} else {

				imm, _ = strconv.ParseUint(immStr, 10, 16)
			}
			imm16 := uint16(imm)

			rom = append(rom, 0x7000|srcRegNum<<8|imm16)
			labeledLine = false
		case "move":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x8000|srcRegNum<<8|destRegNum<<4|0)
			labeledLine = false
		case "or":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x8000|srcRegNum<<8|destRegNum<<4|1)
			labeledLine = false
		case "and":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x8000|srcRegNum<<8|destRegNum<<4|2)
			labeledLine = false
		case "xor":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x8000|srcRegNum<<8|destRegNum<<4|3)
			labeledLine = false
		case "addr":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x8000|srcRegNum<<8|destRegNum<<4|4)
			labeledLine = false
		case "sub":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x8000|srcRegNum<<8|destRegNum<<4|5)
			labeledLine = false
		case "slh":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x8000|srcRegNum<<8|destRegNum<<4|6)
			labeledLine = false

		case "skrne":
			tok = s.Scan()
			srcReg := s.TokenText()

			if !isValidReg(srcReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", srcReg, s.Pos().Line)
				panic(errStr)
			}
			srcRegNum := getRegNum(srcReg)
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			rom = append(rom, 0x9000|srcRegNum<<8|destRegNum<<4|0)
			labeledLine = false
		case "loadi":
			tok = s.Scan()
			immStr := s.TokenText()
			var imm uint64
			if immStr[0] == '0' && immStr[1] == 'x' {
				imm, _ = strconv.ParseUint(immStr[2:], 16, 16)
			} else {

				imm, _ = strconv.ParseUint(immStr, 10, 16)
			}
			imm16 := uint16(imm)
			rom = append(rom, 0xA000|imm16)
			labeledLine = false
		case "jumpi":
			tok = s.Scan()
			immStr := s.TokenText()
			var imm uint64
			if immStr[0] == '0' && immStr[1] == 'x' {
				imm, _ = strconv.ParseUint(immStr[2:], 16, 16)
			} else {

				imm, _ = strconv.ParseUint(immStr, 10, 16)
			}
			imm16 := uint16(imm)
			rom = append(rom, 0xB000|imm16)
			labeledLine = false
		case "rand":
			tok = s.Scan()
			destReg := s.TokenText()

			if !isValidReg(destReg) {
				errStr := fmt.Errorf("invalid register operand %s at line %d", destReg, s.Pos().Line)
				panic(errStr)
			}
			destRegNum := getRegNum(destReg)

			tok = s.Scan()

			immStr := s.TokenText()
			var imm uint64
			if immStr[0] == '0' && immStr[1] == 'x' {
				imm, _ = strconv.ParseUint(immStr[2:], 16, 16)
			} else {

				imm, _ = strconv.ParseUint(immStr, 10, 16)
			}
			imm16 := uint16(imm)

			rom = append(rom, 0xC000|destRegNum<<8|imm16)
			labeledLine = false
		case "addi":
			srcRegNum := parseReg(&s)
			rom = append(rom, 0xF01E|srcRegNum<<8)
			labeledLine = false
		case "stor":
			srcRegNum := parseReg(&s)
			rom = append(rom, 0xF055|srcRegNum<<8)
			labeledLine = false
		case "read":
			srcRegNum := parseReg(&s)
			rom = append(rom, 0xF065|srcRegNum<<8)
			labeledLine = false

		default:
			if labeledLine {
				badInst := fmt.Errorf("unrecognized instruction at line %d", s.Pos().Line)
				panic(badInst)
			}
			backRefs := undefTable[s.TokenText()]
			fmt.Println("backrefs: ", backRefs)
			if len(backRefs) > 0 {
				for _, ref := range backRefs {
					rom[ref] = rom[ref] | (progStart + uint16(s.Pos().Line-1)*2)
				}
			}
			symbolTable[text] = progStart + uint16(s.Pos().Line-1)*2
			currInstNum--
			labeledLine = true
		}
		// fmt.Println(rom)
		currInstNum++
		fmt.Println("sym", symbolTable)
		fmt.Println("undef", undefTable)
	}

	// fmt.Println(rom)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, rom)
	out, _ := os.Create(outFile)
	defer out.Close()
	out.Write(buf.Bytes())

}
