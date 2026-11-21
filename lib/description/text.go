package description

import (
	"regexp"
	"slices"
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

// short forms that end in a full stop without ending the sentence
var shortForms = []string{"e.g.", "i.e.", "etc.", "vs.", "cf.", "approx.", "incl.", "excl.", "viz.", "no.", "min.", "max.", "a.m.", "p.m."}

// Sentences splits masked text into its sentences, each from its first word to its closing mark. A full stop after a short form such as "e.g.",
// an initial or the number of a list item does not end one.
func Sentences(masked string) []Word {
	var sentences []Word
	start := -1
	for _, word := range Words(masked) {
		if start < 0 {
			start = word.Start
		}

		stem := strings.ToLower(strings.Trim(word.Text, "()[]{}\"'*_"))
		if stem == "" || !strings.ContainsRune(".?!", rune(stem[len(stem)-1])) || strings.HasSuffix(stem, "..") {
			continue
		}
		if slices.Contains(shortForms, stem) || len(stem) == 2 || strings.Trim(stem[:len(stem)-1], "0123456789") == "" {
			continue
		}

		sentences = append(sentences, Word{Start: start, End: word.End, Text: masked[start:word.End]})
		start = -1
	}

	if start >= 0 {
		end := len(strings.TrimRight(masked, " \t\n\r"))
		sentences = append(sentences, Word{Start: start, End: end, Text: masked[start:end]})
	}

	return sentences
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
