package main

import (
	"bytes"
	"encoding/binary"
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
	rom := make([]uint16, 0)
	symbolTable := make(map[string]uint16)
	src, _ := ioutil.ReadFile("test.8asm")
	var s scanner.Scanner
	s.Init(bytes.NewBuffer(src))
	labeledLine := false
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		fmt.Printf("pos: %s text: %s\n", s.Position, s.TokenText())
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
			rom = append(rom, 0x1000|symbolTable[s.TokenText()])
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
			fmt.Printf("reg: %#x\n", regNum16)
			// now get the second operand
			tok = s.Scan()
			imm64, _ := strconv.ParseUint(s.TokenText(), 10, 16)
			imm := uint16(imm64)

			rom = append(rom, 0x3000|regNum16<<8|imm)

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
			fmt.Printf("reg: %#x\n", regNum16)
			// now get the second operand
			tok = s.Scan()
			imm64, _ := strconv.ParseUint(s.TokenText(), 10, 16)
			imm := uint16(imm64)

			rom = append(rom, 0x4000|regNum16<<8|imm)

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
		case "addi":
			srcRegNum := parseReg(&s)
			rom = append(rom, 0xF01E|srcRegNum<<8)
		case "stor":
			srcRegNum := parseReg(&s)
			rom = append(rom, 0xF055|srcRegNum<<8)
		case "read":
			srcRegNum := parseReg(&s)
			rom = append(rom, 0xF065|srcRegNum<<8)

		default:
			if labeledLine {
				badInst := fmt.Errorf("unrecognized instruction at line %d", s.Pos().Line)
				panic(badInst)
			}
			symbolTable[text] = progStart + uint16(s.Pos().Line)
			labeledLine = true
		}
		fmt.Println(rom)
	}

	fmt.Println(rom)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, rom)
	out, _ := os.Create("test.bin")
	defer out.Close()
	out.Write(buf.Bytes())

}
