package pcpdf

// splitWords splits on spaces and newlines for word-wrap.
func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\n' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// splitLines splits a string on \n for multi-line text rendering.
func splitLines(s string) []string {
	var lines []string
	cur := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, cur)
			cur = ""
		} else {
			cur += string(c)
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func pdfEscape(s string) string {
	// Unicode code point → WinAnsi (Windows-1252) byte value.
	winansi := map[rune]byte{
		// ── Latin-1 supplement U+00A0–U+00FF (identity in WinAnsi) ──────────
		'\u00A0': 0xA0, '\u00A1': 0xA1, '\u00A2': 0xA2, '\u00A3': 0xA3,
		'\u00A4': 0xA4, '\u00A5': 0xA5, '\u00A6': 0xA6, '\u00A7': 0xA7,
		'\u00A8': 0xA8, '\u00A9': 0xA9, '\u00AA': 0xAA, '\u00AB': 0xAB,
		'\u00AC': 0xAC, '\u00AD': 0xAD, '\u00AE': 0xAE, '\u00AF': 0xAF,
		'\u00B0': 0xB0, '\u00B1': 0xB1, '\u00B2': 0xB2, '\u00B3': 0xB3,
		'\u00B4': 0xB4, '\u00B5': 0xB5, '\u00B6': 0xB6, '\u00B7': 0xB7,
		'\u00B8': 0xB8, '\u00B9': 0xB9, '\u00BA': 0xBA, '\u00BB': 0xBB,
		'\u00BC': 0xBC, '\u00BD': 0xBD, '\u00BE': 0xBE, '\u00BF': 0xBF,
		'\u00C0': 0xC0, '\u00C1': 0xC1, '\u00C2': 0xC2, '\u00C3': 0xC3,
		'\u00C4': 0xC4, '\u00C5': 0xC5, '\u00C6': 0xC6, '\u00C7': 0xC7,
		'\u00C8': 0xC8, '\u00C9': 0xC9, '\u00CA': 0xCA, '\u00CB': 0xCB,
		'\u00CC': 0xCC, '\u00CD': 0xCD, '\u00CE': 0xCE, '\u00CF': 0xCF,
		'\u00D0': 0xD0, '\u00D1': 0xD1, '\u00D2': 0xD2, '\u00D3': 0xD3,
		'\u00D4': 0xD4, '\u00D5': 0xD5, '\u00D6': 0xD6, '\u00D7': 0xD7,
		'\u00D8': 0xD8, '\u00D9': 0xD9, '\u00DA': 0xDA, '\u00DB': 0xDB,
		'\u00DC': 0xDC, '\u00DD': 0xDD, '\u00DE': 0xDE, '\u00DF': 0xDF,
		'\u00E0': 0xE0, '\u00E1': 0xE1, '\u00E2': 0xE2, '\u00E3': 0xE3,
		'\u00E4': 0xE4, '\u00E5': 0xE5, '\u00E6': 0xE6, '\u00E7': 0xE7,
		'\u00E8': 0xE8, '\u00E9': 0xE9, '\u00EA': 0xEA, '\u00EB': 0xEB,
		'\u00EC': 0xEC, '\u00ED': 0xED, '\u00EE': 0xEE, '\u00EF': 0xEF,
		'\u00F0': 0xF0, '\u00F1': 0xF1, '\u00F2': 0xF2, '\u00F3': 0xF3,
		'\u00F4': 0xF4, '\u00F5': 0xF5, '\u00F6': 0xF6, '\u00F7': 0xF7,
		'\u00F8': 0xF8, '\u00F9': 0xF9, '\u00FA': 0xFA, '\u00FB': 0xFB,
		'\u00FC': 0xFC, '\u00FD': 0xFD, '\u00FE': 0xFE, '\u00FF': 0xFF,

		// ── Windows-1252 extras 0x80–0x9F (absent from ISO-8859-1) ──────────
		'\u20AC': 0x80, // €
		'\u201A': 0x82, // ‚
		'\u0192': 0x83, // ƒ
		'\u201E': 0x84, // „
		'\u2026': 0x85, // …  ellipsis
		'\u2020': 0x86, // †
		'\u2021': 0x87, // ‡
		'\u02C6': 0x88, // ˆ
		'\u2030': 0x89, // ‰
		'\u0160': 0x8A, // Š
		'\u2039': 0x8B, // ‹
		'\u0152': 0x8C, // Œ
		'\u017D': 0x8E, // Ž
		'\u2018': 0x91, // '  left single quote
		'\u2019': 0x92, // '  right single quote / apostrophe
		'\u201C': 0x93, // "  left double quote
		'\u201D': 0x94, // "  right double quote
		'\u2022': 0x95, // •  bullet
		'\u2013': 0x96, // –  en dash
		'\u2014': 0x97, // —  em dash  ← was causing ??? bug
		'\u02DC': 0x98, // ˜
		'\u2122': 0x99, // ™
		'\u0161': 0x9A, // š
		'\u203A': 0x9B, // ›
		'\u0153': 0x9C, // œ
		'\u017E': 0x9E, // ž
		'\u0178': 0x9F, // Ÿ
	}

	var out []byte
	for _, r := range s { // range over runes, not bytes
		switch {
		case r == '(' || r == ')' || r == '\\':
			out = append(out, '\\', byte(r))
		case r >= 0x20 && r <= 0x7E:
			out = append(out, byte(r))
		default:
			if b, ok := winansi[r]; ok {
				// Octal escape: \NNN
				out = append(out,
					'\\',
					byte('0'+b/64),
					byte('0'+(b/8)%8),
					byte('0'+b%8),
				)
			} else {
				// Dash-like characters that have no WinAnsi slot → plain hyphen
				switch r {
				case '\u2010', '\u2011', '\u2012', '\u2015',
					'\uFE58', '\uFE63', '\uFF0D':
					out = append(out, '-')
				default:
					out = append(out, '?')
				}
			}
		}
	}
	return string(out)
}
