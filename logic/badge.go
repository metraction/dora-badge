package logic

import "fmt"

const BadgeWarningColor = "#dfb317"
const BadgeSuccessColor = "#4c1"

// BadgeSVG renders an SVG badge with the given text, value, and optional color (defaults to green if empty).
func BadgeSVG(text, value, color string) string {
	if color == "" {
		color = BadgeSuccessColor
	}
	// Estimate width: 7px per char, plus padding
	labelLen := len(text)
	valueLen := len(value)
	labelWidth := 6*labelLen + 9 // px
	valueWidth := 7*valueLen + 9 // px
	if labelWidth < 60 {
		labelWidth = 60
	}

	totalWidth := labelWidth + valueWidth
	// Text positions
	labelX := labelWidth / 2
	valueX := labelWidth + valueWidth/2

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20">
  <linearGradient id="b" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a">
    <rect width="%d" height="20" rx="3" fill="#fff"/>
  </mask>
  <g mask="url(#a)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="%s"/>
    <rect width="%d" height="20" fill="url(#b)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`,
		totalWidth, totalWidth, labelWidth, labelWidth, valueWidth, color, totalWidth,
		labelX, text, labelX, text, valueX, value, valueX, value)
}
