package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"syscall/js"
)

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("WordMatch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return WordMatch(args[0].String(), args[1].String())
	}))
	js.Global().Set("SentenceMatch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return SentenceMatch(args[0].String(), args[1].String())
	}))
	js.Global().Set("TextMatch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return TextMatch(args[0].String(), args[1].String())
	}))
	js.Global().Set("LCS", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return LCS(StrSeq(args[0].String()), StrSeq(args[1].String()))
	}))
	<-c 
}

func SentenceMatchWrapper(sent1, sent2 string) bool {
	return SentenceMatch(sent1, sent2)
}

var sentRe = regexp.MustCompile("[;!.?]+")
var phraseRe = regexp.MustCompile("[:;,!.?]+")
var noAlpha = regexp.MustCompile(`[^a-zA-Z]`)
var noAlphaS = regexp.MustCompile(`[^a-zA-Z\s]`)
var reS = regexp.MustCompile(`\s+`)

func SplitAll(s string, r *regexp.Regexp) []string {
	return r.Split(s, -1)
}
func RemoveAll(s string, r *regexp.Regexp) string {
	return r.ReplaceAllString(s, "")
}

func eq[T comparable](a, b T) bool {
	return !(a != b)
}

func LCS[T comparable](seq1, seq2 []T, compare ...func(a, b T) bool) int {
	var comp func(a, b T) bool
	if len(compare) > 0 {
		comp = compare[0]
	} else {
		comp = eq
	}
	arr1 := seq1
	arr2 := seq2
	if len(arr2) > len(arr1) {
		arr2 = seq1
		arr1 = seq2
	}
	arr1len := len(arr1)
	arr2len := len(arr2)
	dp_len := arr1len + 1
	dpi_len := arr2len + 1
	dp := make([][]int, arr1len+1)
	for i := 0; i != dp_len; i++ {
		dp[i] = make([]int, dpi_len)
	}
	for i := 1; i != dp_len; i++ {
		for x := 1; x != dpi_len; x++ {
			if arr1[i-1] == arr2[x-1] || comp(arr1[i-1], arr2[x-1]) {
				dp[i][x] = dp[i-1][x-1] + 1
			} else {
				dp[i][x] = max(dp[i][x-1], dp[i-1][x])
			}
		}
	}
	return dp[arr1len][arr2len]
}

func SeqMatch[T comparable](seq1, seq2 []T, compare ...func(a, b T) bool) bool {
	lcs := LCS(seq1, seq2, compare...)
	maxlen := math.Max(float64(len(seq1)), float64(len(seq2)))
	return lcs >= int(math.Floor(maxlen*0.8))
}

func StrSeq(word string) []string {
	return strings.Split(word, "")
}

func StrMatch(a, b string, compare ...func(a, b string) bool) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return SeqMatch(StrSeq(a), StrSeq(b))
}

var wordMap = make(map[string]bool)

func WordMatch(word1, word2 string) bool {
	word1 = strings.ToLower(word1)
	word1 = RemoveAll(word1, noAlpha)
	word2 = strings.ToLower(word2)
	word2 = RemoveAll(word2, noAlpha)
	if len(word1) == 0 || len(word2) == 0 {
		return false
	}
	if eq(word1, word2) {
		return true
	}
	if word1 > word2 {
		word1, word2 = word2, word1
	}
	key := word1 + ":" + word2
	if value, ok := wordMap[key]; ok {
		return value
	}
	value := StrMatch(word1, word2)
	wordMap[key] = value
	return value
}

func WordSeq(sent string) []string {
	return SplitAll(sent, reS)
}

func SentenceMatch(sent1, sent2 string) bool {
	if len(sent1) == 0 || len(sent2) == 0 {
		return false
	}
	sent1 = strings.ToLower(sent1)
	sent2 = strings.ToLower(sent2)
	return SeqMatch(WordSeq(sent1), WordSeq(sent2), WordMatch)
}

func TextMatch(text1, text2 string) bool {
	t1 := SplitAll(text1, sentRe)
	t2 := SplitAll(text2, sentRe)
	if SeqMatch(t1, t2, SentenceMatch) {
		return true
	}
	t1 = SplitAll(text1, phraseRe)
	t2 = SplitAll(text2, phraseRe)
	if SeqMatch(t1, t2, SentenceMatch) {
		return true
	}
	t1 = SplitAll(text1, noAlphaS)
	t2 = SplitAll(text2, noAlphaS)
	return SeqMatch(t1, t2, SentenceMatch)
}

func LangMatch(lang1, lang2 string) bool {
	return WordMatch(lang1, lang2) || SentenceMatch(lang1, lang2) || TextMatch(lang1, lang2)
}

func AnyMatch[T any, U any](any1 T, any2 U) bool {
	return LangMatch(fmt.Sprint(any1), fmt.Sprint(any2))
}
