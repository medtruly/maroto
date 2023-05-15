package internal

import (
	"bufio"
	"strings"

	"github.com/johnfercher/maroto/internal/fpdf"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/props"
)

// Text is the abstraction which deals of how to add text inside PDF.
type Text interface {
	Add(text string, cell Cell, textProp props.Text)
	GetLinesQuantity(text string, fontFamily props.Text, colWidth float64) int
}

type text struct {
	pdf  fpdf.Fpdf
	math Math
	font Font
}

// NewText create a Text.
func NewText(pdf fpdf.Fpdf, math Math, font Font) *text {
	return &text{
		pdf,
		math,
		font,
	}
}

// Add a text inside a cell.
func (s *text) Add(text string, cell Cell, textProp props.Text) {
	s.font.SetFont(textProp.Family, textProp.Style, textProp.Size)

	originalColor := s.font.GetColor()
	s.font.SetColor(textProp.Color)

	// duplicated
	_, _, fontSize := s.font.GetFont()
	fontHeight := fontSize / s.font.GetScaleFactor()

	cell.Y += fontHeight

	// Apply Unicode before calc spaces
	unicodeText := s.textToUnicode(text, textProp)
	accumulateOffsetY := 0.0

	lines := s.getLines(unicodeText, cell.Width, textProp.Extrapolate)

	for index, line := range lines {
		lineWidth := s.pdf.GetStringWidth(line)
		_, _, fontSize := s.font.GetFont()
		textHeight := fontSize / s.font.GetScaleFactor()

		s.addLine(textProp, cell.X, cell.Width, cell.Y+float64(index)*textHeight+accumulateOffsetY, lineWidth, line)
		accumulateOffsetY += textProp.VerticalPadding
	}

	s.font.SetColor(originalColor)
}

// GetLinesQuantity retrieve the quantity of lines which a text will occupy to avoid that text to extrapolate a cell.
func (s *text) GetLinesQuantity(text string, textProp props.Text, colWidth float64) int {
	translator := s.pdf.UnicodeTranslatorFromDescriptor("")
	s.font.SetFont(textProp.Family, textProp.Style, textProp.Size)

	// Apply Unicode.
	textTranslated := translator(text)

	lines := s.getLines(textTranslated, colWidth, textProp.Extrapolate)
	return len(lines)
}

func (s *text) getLines(text string, colWidth float64, extrapolate bool) []string {

	var lines []string
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := sc.Text()

		if s.pdf.GetStringWidth(line) < colWidth || extrapolate {
			lines = append(lines, line)
			continue
		}

		currentlySize := 0.0
		var currentlyLine string
		words := strings.Split(line, " ")
		for i, word := range words {
			if i != len(words)-1 {
				// if not last word, add space
				word += " "
			}
			wordWidth := s.pdf.GetStringWidth(word)
			if wordWidth > colWidth {
				// Single word is too long
				lines = append(lines, word)
				continue
			}

			if wordWidth+currentlySize < colWidth {
				currentlyLine += word
				currentlySize += wordWidth
			} else {
				// complete line and add word to new line
				lines = append(lines, currentlyLine)
				currentlyLine = word
				currentlySize = wordWidth
			}
		}
		if currentlyLine != "" {
			// take care of the tail line
			lines = append(lines, currentlyLine)
		}
	}

	return lines
}

func (s *text) addLine(textProp props.Text, xColOffset, colWidth, yColOffset, textWidth float64, text string) {
	left, top, _, _ := s.pdf.GetMargins()

	if textProp.Align == consts.Left {
		s.pdf.Text(xColOffset+left, yColOffset+top, text)
		return
	}

	var modifier float64 = 2

	if textProp.Align == consts.Right {
		modifier = 1
	}

	dx := (colWidth - textWidth) / modifier

	s.pdf.Text(dx+xColOffset+left, yColOffset+top, text)
}

func (s *text) textToUnicode(txt string, props props.Text) string {
	if props.Family == consts.Arial ||
		props.Family == consts.Helvetica ||
		props.Family == consts.Symbol ||
		props.Family == consts.ZapBats ||
		props.Family == consts.Courier {
		translator := s.pdf.UnicodeTranslatorFromDescriptor("")
		return translator(txt)
	}

	return txt
}
