package description

import (
	"regexp"
	"strings"
)

// Blank is the byte Mask puts where the text is not prose.
const Blank = '\x01'

var (
	code       = regexp.MustCompile("`[^`]*`")
	linkTarget = regexp.MustCompile(`\]\((?:[^()\s]|\([^()\s]*\))*\)`)
	address    = regexp.MustCompile(`<https?://[^>\s]*>|https?://[^\s)\]>"']+|\b[\w-]+(?:\.[\w-]+)+/[^\s)\]>"']*`)
)

// Mask returns the text with everything that is not prose blanked out byte for byte, so a place in the result is the same place in the text:
// code in backticks, where a link points, and web addresses.
func Mask(text string) string {
	masked := []byte(text)
	blank := func(start, end int) {
		for i := start; i < end; i++ {
			masked[i] = Blank
		}
	}

	for _, at := range code.FindAllStringIndex(text, -1) {
		blank(at[0], at[1])
	}

	for _, at := range linkTarget.FindAllIndex(masked, -1) {
		blank(at[0]+1, at[1])
	}

	// a full stop or comma straight after an address ends the sentence, not the address
	for _, at := range address.FindAllIndex(masked, -1) {
		end := at[1]
		if masked[at[0]] != '<' {
			end = at[0] + len(strings.TrimRight(string(masked[at[0]:end]), ".,;:?!"))
		}
		blank(at[0], end)
	}

	return string(masked)
}

// Word is a run of text with no space in it, and where it sits.
type Word struct {
	Start, End int
	Text       string
}

// Words splits masked text at its spaces.
func Words(masked string) []Word {
	var words []Word
	start := -1
	for i := range len(masked) {
		if !strings.ContainsRune(" \t\n\r", rune(masked[i])) {
			if start < 0 {
				start = i
			}
			continue
		}

		if start >= 0 {
			words = append(words, Word{Start: start, End: i, Text: masked[start:i]})
			start = -1
		}
	}

	if start >= 0 {
		words = append(words, Word{Start: start, End: len(masked), Text: masked[start:]})
	}

	return words
}

// Trim returns the word without the brackets, quotes and emphasis marks opening it, and without any of the given characters closing it.
func (w Word) Trim(closing string) Word {
	text := strings.TrimLeft(w.Text, "([{\"'*_")
	w.Start = w.End - len(text)
	text = strings.TrimRight(text, closing)
	w.End = w.Start + len(text)
	w.Text = text

	return w
}
