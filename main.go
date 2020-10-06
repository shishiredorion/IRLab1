package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"
	"unicode"
)

type trie struct {
	isEnd bool
	next [256]*trie
}

func (t *trie) exist(c byte) bool {
	return t.next[c] != nil
}

func (t *trie) forceToNext(c byte) *trie {
	if !t.exist(c){
		t.next[c] = &trie{}
	}
	return t.next[c]
}

type WordDict struct {
	dict *trie    // 字典,实则为一颗树
	wordCount int // 字典中含有词数
	maxLen int    // 最大长度
}

func (wordDict *WordDict)ReadDict()  {
	f, err := os.Open("corpus.dict.txt")
	if err != nil{
		log.Fatalln(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	wordDict.parseHeader(reader)
	wordDict.dict = &trie{}
	wordDict.addWord(reader)
}

func (wordDict *WordDict)parseHeader(reader *bufio.Reader)  {
	header, err := reader.ReadBytes('\n')
	header = header[:len(header) - 1]
	if err != nil || len(header) < 3{
		log.Fatalln("Header Format Error")
	}

	gapIdx := 0
	for gapIdx < len(header){
		if header[gapIdx] == 9{
			break
		}
		gapIdx++
	}
	if gapIdx == 0 || gapIdx >= len(header) - 1{
		log.Fatalln("Header Format Error")
	}
	wordDict.wordCount, wordDict.maxLen = readNum(header[:gapIdx]), readNum(header[gapIdx + 1:])
}

func (wordDict *WordDict)addWord(reader *bufio.Reader)  {
	for i := 0; i < wordDict.wordCount; i++{
		line, isEOF := readLine(reader)
		if isEOF{
			break
		}

		temp, count, i := wordDict.dict, 0, 0
		for i < len(line){
			if count > wordDict.maxLen{
				log.Fatalln("The word's length overflow.")
			}

			utf8CharacterLen := calculateUTF8CharacterByteLen(line[i])
			temp = temp.forceToNext(line[i])
			if utf8CharacterLen == 1{
				i++
				continue
			}

			checkUTF8ByteValid(line, i, utf8CharacterLen)
			for j := 1; j < utf8CharacterLen; j++{
				temp = temp.forceToNext(line[i + j])
			}
			i = i + utf8CharacterLen
			count++
		}
		temp.isEnd = true
	}
}

func (wordDict *WordDict)SegmentAndCalculatePRF()  {
	if wordDict.dict == nil{
		log.Fatalln("Word Dict Not exist")
	}
	buffer := wordDict.segment()
	wordDict.calculatePRF(buffer)
	wordDict.output(buffer)
}

func (wordDict *WordDict)segment() (buffer [][]string) {
	sentenceFile, err := os.Open("corpus.sentence.txt")
	if err != nil{
		log.Fatalln(err)
	}
	defer sentenceFile.Close()

	reader := bufio.NewReader(sentenceFile)

	l := 0
	for {
		line, isEOF := readLine(reader)
		if isEOF{
			break
		}

		i := 0
		buffer = append(buffer, []string{})
		for i < len(line){
			var isBreakdown bool
			temp, j, lastEnd, count := wordDict.dict, i, i, 0
			for j < len(line){
				if count > wordDict.maxLen{
					break
				}

				j, lastEnd, temp, isBreakdown = getUTF8CharacterWithDict(line, j, lastEnd, temp)
				if isBreakdown{
					break
				}
				count++
			}
			if lastEnd == i{
				k := getUTF8CharacterWithoutDict(line, i)
				buffer[l] = append(buffer[l], string(line[i: k]))
				i = k
				continue
			}
			i, buffer[l] = lastEnd, append(buffer[l], string(line[i:lastEnd]))
		}
		l++
	}
	return
}

func (wordDict *WordDict)calculatePRF(buffer [][]string)  {
	ansFile, err := os.Open("corpus.answer.txt")
	if err != nil{
		log.Fatalln(err)
	}
	defer ansFile.Close()

	ansReader := bufio.NewReader(ansFile)
	prfFile, err1 := os.Create("corpus.evaluation.txt")
	if err1 != nil{
		log.Fatalln(err1)
	}

	l := 0
	for {
		line, err := ansReader.ReadString('\n')
		if err == io.EOF{
			break
		}else if err != nil{
			log.Fatalln(err)
		}
		line = line[:len(line) - 1]

		p, r, f := calculatePRF(strings.Fields(line), buffer[l])
		_, err2 := prfFile.WriteString(fmt.Sprintf("%.3f, %.3f, %.3f\n", p, r, f))
		if err2 != nil{
			log.Fatalln(err2)
		}
		l++
	}
}

func (wordDict *WordDict) output(buffer [][]string) {
	outputFile, err := os.Create("corpus.output.txt")
	if err != nil{
		log.Fatalln(err)
	}
	for i := 0; i < len(buffer); i++{
		_, err := outputFile.WriteString(strings.Join(buffer[i], " ") + " \n")
		if err != nil{
			log.Fatalln(err)
		}
	}
}

func readNum(arr []byte) int {
	num := 0
	for i := 0; i < len(arr); i++{
		if !unicode.IsDigit(rune(arr[i])){
			log.Fatalln("Read Num Error")
		}
		if num > math.MaxInt32 / 10 || num == math.MaxInt32 / 10 && int(arr[i]) > 55 {
			log.Fatalln("The Num OverFlow")
		}
		num = num * 10 + int(arr[i]) - 48
	}
	return num
}

func readLine(reader *bufio.Reader) ([]byte, bool) {
	line, err := reader.ReadBytes(0xA)
	if err == io.EOF{
		return nil, true
	}else if err != nil{
		log.Fatalln(err)
	}
	line = line[:len(line) - 1]
	return line, false
}

func calculateUTF8CharacterByteLen(c byte) int {
	if c & 0x80 == 0{
		return 1
	}

	multiBytesUtf8CharacterLen := 0
	for c & 0x80 != 0{
		multiBytesUtf8CharacterLen++
		c = c << 1
	}
	if multiBytesUtf8CharacterLen > 4 || multiBytesUtf8CharacterLen < 2{
		log.Fatalln("The UTF-8 Character Format Error")
	}
	return multiBytesUtf8CharacterLen
}

func checkUTF8ByteValid(line []byte, i int, utf8CharacterLen int) {
	for k := 1; k < utf8CharacterLen; k++ {
		if i+k >= len(line) || line[i+k] < 128 || line[i+k] > 191 {
			log.Fatalln("The UTF-8 Character Format Error")
		}
	}
}

func getUTF8CharacterWithDict(line []byte, i int, lastEnd int, t *trie) (int, int, *trie, bool) {
	if t.next[line[i]] == nil{
		return i, lastEnd, nil, true
	}
	t = t.next[line[i]]

	utf8CharacterLen := calculateUTF8CharacterByteLen(line[i])
	if utf8CharacterLen == 1{
		if t.isEnd{
			return i + 1, i + 1, t, false
		}
		return i + 1, lastEnd, t, false
	}
	checkUTF8ByteValid(line, i, utf8CharacterLen)
	for k := 1; k < utf8CharacterLen; k++{
		if !t.exist(line[i + k]){
			return i, lastEnd, nil, true
		}
		t = t.next[line[i + k]]
	}

	if t.isEnd{
		return i + utf8CharacterLen, i + utf8CharacterLen, t, false
	}
	return i + utf8CharacterLen, lastEnd, t, false
}

func getUTF8CharacterWithoutDict(line []byte, i int) int {
	utf8CharacterLen := calculateUTF8CharacterByteLen(line[i])
	if utf8CharacterLen > 1{
		checkUTF8ByteValid(line, i, utf8CharacterLen)
	}
	return i + utf8CharacterLen
}

func calculatePRF(ansSegmentation, mySegmentation []string) (float64, float64, float64) {
	count := float64(correctlySegmentedCount(ansSegmentation, mySegmentation))
	p, r := count / float64(len(mySegmentation)), count / float64(len(ansSegmentation))
	return p, r, p * r * 2 / (p + r)
}

func correctlySegmentedCount(ansSegmentation []string, mySegmentation []string) int {
	count, i, j, l1, l2 := 0, 0, 0, 0, 0
	for i < len(ansSegmentation) && j < len(mySegmentation){
		if l1 == l2{
			if len(ansSegmentation[i]) == len(mySegmentation[j]){
				count++
			}
			l1, l2, i, j = l1 + len(ansSegmentation[i]), l2 + len(mySegmentation[j]), i + 1, j + 1
		}else if l1 > l2{
			l2, j = l2 + len(mySegmentation[j]), j + 1
		}else if l2 > l1{
			l1, i = l1 + len(ansSegmentation[i]), i + 1
		}
	}
	return count
}

func main()  {
	a := time.Now()
	wordDict := WordDict{}
	wordDict.ReadDict()
	wordDict.SegmentAndCalculatePRF()
	fmt.Println("用时", time.Now().Sub(a))
}
