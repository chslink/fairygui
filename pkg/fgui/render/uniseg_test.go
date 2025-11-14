package render

import (
	"testing"

	"github.com/rivo/uniseg"
)

func TestGraphemeClusters(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			input:    "Hello",
			expected: 5,
		},
		{
			input:    "🇩🇪", // 德国国旗 emoji
			expected: 1, // 应该作为单个字素集群
		},
		{
			input:    "🏳️‍🌈", // 彩虹旗 emoji with zero width joiner
			expected: 1, // 多个码点组成一个字素集群
		},
		{
			input:    "Käse", // 德语，a + diaeresis
			expected: 4, // K, a+ diaeresis, s, e
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := uniseg.GraphemeClusterCount(test.input)
			if result != test.expected {
				t.Errorf("GraphemeClusterCount(%q) = %d; want %d", test.input, result, test.expected)
			}

			// 验证我们的缓存函数
			cached := getGraphemeClusters(test.input)
			if len(cached) != test.expected {
				t.Errorf("getGraphemeClusters(%q) returned %d clusters; want %d", test.input, len(cached), test.expected)
			}
		})
	}
}

func TestIsBreakRune(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{" ", true},
		{"\n", true},
		{"\t", true},
		{"\r", true},
		{"a", false},
		{"中", false},
		{"🇩🇪", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			runes := []rune(test.input)
			if len(runes) == 0 {
				t.Skip("empty input")
			}
			result := isBreakRune(runes[0])
			if result != test.expected {
				t.Errorf("isBreakRune(%q) = %v; want %v", test.input, result, test.expected)
			}
		})
	}
}

func TestSkipWhitespaceGraphemes(t *testing.T) {
	tests := []struct {
		input    string
		start    int
		expected int
	}{
		{
			input:    "  Hello",
			start:    0,
			expected: 2, // 跳过两个空格
		},
		{
			input:    "Hello",
			start:    0,
			expected: 0, // 没有空格
		},
		{
			input:    "  \n  ",
			start:    0,
			expected: 5, // 跳过所有空白（共5个grapheme集群）
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := skipWhitespaceGraphemes(test.input, test.start)
			if result != test.expected {
				t.Errorf("skipWhitespaceGraphemes(%q, %d) = %d; want %d", test.input, test.start, result, test.expected)
			}
		})
	}
}

func TestGraphemeWidth(t *testing.T) {
	// 这个测试需要字体环境，所以我们只测试数据结构
	t.Run("cache", func(t *testing.T) {
		input := "Hello"
		clusters1 := getGraphemeClusters(input)
		clusters2 := getGraphemeClusters(input)

		// 应该返回相同长度的切片（缓存）
		if len(clusters1) != len(clusters2) {
			t.Errorf("Expected same cluster count, got %d vs %d", len(clusters1), len(clusters2))
		}
	})
}
