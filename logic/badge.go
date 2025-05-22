package logic

import "fmt"

func BadgeSVG(text, value string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="200" height="20">
  <linearGradient id="b" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a">
    <rect width="200" height="20" rx="3" fill="#fff"/>
  </mask>
  <g mask="url(#a)">
    <rect width="140" height="20" fill="#555"/>
    <rect x="140" width="60" height="20" fill="#4c1"/>
    <rect width="200" height="20" fill="url(#b)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="70" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="70" y="14">%s</text>
    <text x="170" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="170" y="14">%s</text>
  </g>
</svg>`, text, text, value, value)
}
