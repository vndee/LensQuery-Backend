package model

type MathpixOptions struct {
	MathInlineDelimiters []string `json:"math_inline_delimiters"`
	RmSpaces             bool     `json:"rm_spaces"`
}
