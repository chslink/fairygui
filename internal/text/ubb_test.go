package textutil

import (
	"testing"
)

func TestParseUBB(t *testing.T) {
	base := Style{
		Color:     "#ffffff",
		Bold:      false,
		Italic:    false,
		Underline: false,
		Font:      "",
		FontSize:  16,
	}
	cases := []struct {
		name     string
		input    string
		expected []Segment
	}{
		{
			name:  "plain",
			input: "hello",
			expected: []Segment{
				{Text: "hello", Style: base},
			},
		},
		{
			name:  "bold_and_color",
			input: "[b][color=#ff0000]hi[/color][/b]",
			expected: []Segment{
				{Text: "hi", Style: Style{Color: "#ff0000", Bold: true, Italic: false, Underline: false, Font: "", FontSize: 16}},
			},
		},
		{
			name:  "nested_size_font",
			input: "[size=24][font=ui://foo]ABC[/font][/size]",
			expected: []Segment{
				{Text: "ABC", Style: Style{Color: "#ffffff", Bold: false, Italic: false, Underline: false, Font: "ui://foo", FontSize: 24}},
			},
		},
		{
			name:  "underline_url",
			input: "[url=event:a]link[/url]",
			expected: []Segment{
				{Text: "link", Style: Style{Color: "#ffffff", Bold: false, Italic: false, Underline: true, Font: "", FontSize: 16}, Link: "event:a"},
			},
		},
		{
			name:  "unknown_tag_literal",
			input: "[foo]bar[/foo]",
			expected: []Segment{
				{Text: "[foo]bar[/foo]", Style: base},
			},
		},
		{
			name:  "line_breaks",
			input: "a[br]b",
			expected: []Segment{
				{Text: "a", Style: base},
				{Text: "\n", Style: base},
				{Text: "b", Style: base},
			},
		},
		{
			name:  "unclosed_tag",
			input: "[b]text",
			expected: []Segment{
				{Text: "text", Style: Style{Color: "#ffffff", Bold: true, Italic: false, Underline: false, Font: "", FontSize: 16}},
			},
		},
		{
			name:  "img_tag_with_content",
			input: "[img]ui://9leh0eyfrpmb6[/img]",
			expected: []Segment{
				{Text: "\uFFFD", Style: base, ImageURL: "ui://9leh0eyfrpmb6"},
			},
		},
		{
			name:  "img_tag_with_attr",
			input: "[img=ui://test/image]",
			expected: []Segment{
				{Text: "\uFFFD", Style: base, ImageURL: "ui://test/image"},
			},
		},
		{
			name:  "text_with_img",
			input: "Hello [img]ui://test/icon[/img] World",
			expected: []Segment{
				{Text: "Hello ", Style: base},
				{Text: "\uFFFD", Style: base, ImageURL: "ui://test/icon"},
				{Text: " World", Style: base},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseUBB(tc.input, base)
			if len(got) != len(tc.expected) {
				t.Fatalf("segment count mismatch: got %d expected %d", len(got), len(tc.expected))
			}
			for i := range got {
				if got[i].Text != tc.expected[i].Text {
					t.Fatalf("segment %d text mismatch: got %q expected %q", i, got[i].Text, tc.expected[i].Text)
				}
				if got[i].Style.Color != tc.expected[i].Style.Color {
					t.Fatalf("segment %d color mismatch: got %q expected %q", i, got[i].Style.Color, tc.expected[i].Style.Color)
				}
				if got[i].Style.Bold != tc.expected[i].Style.Bold {
					t.Fatalf("segment %d bold mismatch: got %v expected %v", i, got[i].Style.Bold, tc.expected[i].Style.Bold)
				}
				if got[i].Style.Italic != tc.expected[i].Style.Italic {
					t.Fatalf("segment %d italic mismatch: got %v expected %v", i, got[i].Style.Italic, tc.expected[i].Style.Italic)
				}
				if got[i].Style.Underline != tc.expected[i].Style.Underline {
					t.Fatalf("segment %d underline mismatch: got %v expected %v", i, got[i].Style.Underline, tc.expected[i].Style.Underline)
				}
				if got[i].Style.Font != tc.expected[i].Style.Font {
					t.Fatalf("segment %d font mismatch: got %q expected %q", i, got[i].Style.Font, tc.expected[i].Style.Font)
				}
				if got[i].Style.FontSize != tc.expected[i].Style.FontSize {
					t.Fatalf("segment %d size mismatch: got %d expected %d", i, got[i].Style.FontSize, tc.expected[i].Style.FontSize)
				}
				if got[i].Link != tc.expected[i].Link {
					t.Fatalf("segment %d link mismatch: got %q expected %q", i, got[i].Link, tc.expected[i].Link)
				}
				if got[i].ImageURL != tc.expected[i].ImageURL {
					t.Fatalf("segment %d imageURL mismatch: got %q expected %q", i, got[i].ImageURL, tc.expected[i].ImageURL)
				}
			}
		})
	}
}
