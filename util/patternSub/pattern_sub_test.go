package patternSub

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPatternProcessor(t *testing.T) {
	type TestCase struct {
		config   map[string]string
		expected map[string]subMap
	}

	testCases := []TestCase{
		{
			// no common prefixes
			config: map[string]string{
				"a.a.a.": "111.",
				"b.b.b.": "222.",
				"c.c.c.": "333.",
				"d.d.d.": "444.",
				"e.e.e.": "555.",
			},
			expected: map[string]subMap{
				// only default rewrite map
				"": {
					"a.a.a.": "111.",
					"b.b.b.": "222.",
					"c.c.c.": "333.",
					"d.d.d.": "444.",
					"e.e.e.": "555.",
				},
			},
		},
		{
			config: map[string]string{
				// #1 - common prefix is `a.a`
				"a.a.1.": "111.",
				"a.a.2.": "222.",
				"a.a.3.": "333.",
				// #2 - common prefix is `a.b`
				"a.b.1.": "111.",
				"a.b.2.": "222.",
				"a.b.3.": "333.",
				"a.b.4.": "444.",
				// #3 - common prefix is `b.b`
				"b.b.1.": "111.",
				"b.b.2.": "222.",
			},
			expected: map[string]subMap{
				// all together
				"": {
					"a.a.1.": "111.",
					"a.a.2.": "222.",
					"a.a.3.": "333.",
					"a.b.1.": "111.",
					"a.b.2.": "222.",
					"a.b.3.": "333.",
					"a.b.4.": "444.",
					"b.b.1.": "111.",
					"b.b.2.": "222.",
				},
				// #1
				"a.a.*.": {
					"a.a.1.": "111.",
					"a.a.2.": "222.",
					"a.a.3.": "333.",
				},
				// #2
				"a.b.*.": {
					"a.b.1.": "111.",
					"a.b.2.": "222.",
					"a.b.3.": "333.",
					"a.b.4.": "444.",
				},
				// #3
				"b.b.*.": {
					"b.b.1.": "111.",
					"b.b.2.": "222.",
				},
			},
		},
	}

	for _, testCase := range testCases {
		patternProcessor := NewPatternProcessor(testCase.config)
		assert.Equal(t, testCase.expected, patternProcessor.prefix)
	}
}

func TestPatternProcessor_ReplacePrefixFunctionArg(t *testing.T) {
	type TestCase struct {
		config   map[string]string
		pattern  string
		expected []SubstituteInfo
	}

	testCases := []TestCase{
		// no replacement
		{
			config: map[string]string{
				"prefix1.from1.": "prefix1.to1.",
				"prefix2.from2.": "prefix2.to2.",
			},
			pattern: "seriesByTag('tag1=value1')",
			expected: []SubstituteInfo{{
				MetricSrc:  "seriesByTag('tag1=value1')",
				MetricDst:  "seriesByTag('tag1=value1')",
				isReplaced: false,
				prefixSrc:  "",
				prefixDst:  "",
				tagInfo:    nil,
			}},
		},
		{
			config: map[string]string{
				"prefix1.from1.": "prefix1.to1.",
				"prefix2.from2.": "prefix2.to2.",
			},
			pattern: "seriesByTag('tag1=value1', 'tag2=value2', 'tag3=value3')",
			expected: []SubstituteInfo{{
				MetricSrc:  "seriesByTag('tag1=value1','tag2=value2','tag3=value3')",
				MetricDst:  "seriesByTag('tag1=value1','tag2=value2','tag3=value3')",
				isReplaced: false,
				prefixSrc:  "",
				prefixDst:  "",
				tagInfo:    nil,
			}},
		},
		{
			config: map[string]string{
				"prefix1.from1.": "prefix1.to1.",
				"prefix2.from2.": "prefix2.to2.",
			},
			pattern: "seriesByTag('name=a.b.c.d')",
			expected: []SubstituteInfo{{
				MetricSrc:  "seriesByTag('name=a.b.c.d')",
				MetricDst:  "seriesByTag('name=a.b.c.d')",
				isReplaced: false,
				prefixSrc:  "",
				prefixDst:  "",
				tagInfo: &nameTagInfo{
					index: 0,
					sign:  "=",
				},
			}},
		},

		// replacement with strict prefix matching
		{
			config: map[string]string{
				"prefix1.from1.": "prefix1.to1.",
				"prefix2.from2.": "prefix2.to2.",
			},
			pattern: "seriesByTag('name=prefix2.from2.something.else')",
			expected: []SubstituteInfo{{
				MetricSrc:  "seriesByTag('name=prefix2.from2.something.else')",
				MetricDst:  "seriesByTag('name=prefix2.to2.something.else')",
				isReplaced: true,
				prefixSrc:  "prefix2.from2.",
				prefixDst:  "prefix2.to2.",
				tagInfo: &nameTagInfo{
					index: 0,
					sign:  "=",
				},
			}},
		},
		{
			config: map[string]string{
				"prefix1.from1.": "prefix1.to1.",
				"prefix2.from2.": "prefix2.to2.",
			},
			pattern: "seriesByTag('tag1=value1', 'tag2=value2', 'name!=prefix1.from1.something.else', 'tag4=value4')",
			expected: []SubstituteInfo{{
				MetricSrc:  "seriesByTag('tag1=value1','tag2=value2','name!=prefix1.from1.something.else','tag4=value4')",
				MetricDst:  "seriesByTag('tag1=value1','tag2=value2','name!=prefix1.to1.something.else','tag4=value4')",
				isReplaced: true,
				prefixSrc:  "prefix1.from1.",
				prefixDst:  "prefix1.to1.",
				tagInfo: &nameTagInfo{
					index: 2,
					sign:  "!=",
				},
			}},
		},
		{
			config: map[string]string{
				"prefix1.from1.": "prefix1.to1.",
				"prefix2.from2.": "",
			},
			pattern: "seriesByTag('name=~prefix2.from2.something.else', 'tag1=value1')",
			expected: []SubstituteInfo{{
				MetricSrc:  "seriesByTag('name=~prefix2.from2.something.else','tag1=value1')",
				MetricDst:  "seriesByTag('name=~something.else','tag1=value1')",
				isReplaced: true,
				prefixSrc:  "prefix2.from2.",
				prefixDst:  "",
				tagInfo: &nameTagInfo{
					index: 0,
					sign:  "=~",
				},
			}},
		},

		// replacement with pattern prefix matching
		{
			config: map[string]string{
				"a.b.c.d.from1.": "a.b.c.d.to1.",
				"a.b.c.d.from2.": "a.b.c.d.to2.",
				"a.b.c.d.from3.": "a.b.c.d.to3.",
			},
			pattern: "seriesByTag('tag1=value1', 'name!=~a.b.c.d.*.')",
			expected: []SubstituteInfo{
				{
					MetricSrc:  "seriesByTag('tag1=value1','name!=~a.b.c.d.from1.')",
					MetricDst:  "seriesByTag('tag1=value1','name!=~a.b.c.d.to1.')",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from1.",
					prefixDst:  "a.b.c.d.to1.",
					tagInfo: &nameTagInfo{
						index: 1,
						sign:  "!=~",
					},
				},
				{
					MetricSrc:  "seriesByTag('tag1=value1','name!=~a.b.c.d.from2.')",
					MetricDst:  "seriesByTag('tag1=value1','name!=~a.b.c.d.to2.')",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from2.",
					prefixDst:  "a.b.c.d.to2.",
					tagInfo: &nameTagInfo{
						index: 1,
						sign:  "!=~",
					},
				},
				{
					MetricSrc:  "seriesByTag('tag1=value1','name!=~a.b.c.d.from3.')",
					MetricDst:  "seriesByTag('tag1=value1','name!=~a.b.c.d.to3.')",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from3.",
					prefixDst:  "a.b.c.d.to3.",
					tagInfo: &nameTagInfo{
						index: 1,
						sign:  "!=~",
					},
				},
			},
		},
		{
			config: map[string]string{
				"prefix.from1.": "prefix.to1.",
				"prefix.from2.": "prefix.to2.",
			},
			pattern: "seriesByTag('name=prefix.*.something.else')",
			expected: []SubstituteInfo{
				{
					MetricSrc:  "seriesByTag('name=prefix.from1.something.else')",
					MetricDst:  "seriesByTag('name=prefix.to1.something.else')",
					isReplaced: true,
					prefixSrc:  "prefix.from1.",
					prefixDst:  "prefix.to1.",
					tagInfo: &nameTagInfo{
						index: 0,
						sign:  "=",
					},
				},
				{
					MetricSrc:  "seriesByTag('name=prefix.from2.something.else')",
					MetricDst:  "seriesByTag('name=prefix.to2.something.else')",
					isReplaced: true,
					prefixSrc:  "prefix.from2.",
					prefixDst:  "prefix.to2.",
					tagInfo: &nameTagInfo{
						index: 0,
						sign:  "=",
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		patternProcessor := NewPatternProcessor(testCase.config)
		substitutes := patternProcessor.ReplacePrefix(testCase.pattern)

		// assert equality disregarding elements order
		assert.ElementsMatch(t, testCase.expected, substitutes, fmt.Sprintf("Test case #%d", i+1))
	}
}

func TestPatternProcessor_ReplacePrefixSimplePattern(t *testing.T) {
	type TestCase struct {
		config   map[string]string
		pattern  string
		expected []SubstituteInfo
	}

	testCases := []TestCase{
		// no replacement
		{
			config: map[string]string{
				"prefix.from.":  "prefix.to.",
				"prefix1.from.": "prefix2.to.",
			},
			pattern: "a.b.c.d",
			expected: []SubstituteInfo{{
				MetricSrc:  "a.b.c.d",
				MetricDst:  "a.b.c.d",
				isReplaced: false,
				prefixSrc:  "",
				prefixDst:  "",
			}},
		},

		//
		// replacement with strict prefix matching
		//

		// replace prefix
		{
			config: map[string]string{
				"prefix1.from.": "prefix2.to.",
				"aaa.":          "111.",
				"bbb.":          "222.",
			},
			pattern: "prefix1.from.something.else",
			expected: []SubstituteInfo{{
				MetricSrc:  "prefix1.from.something.else",
				MetricDst:  "prefix2.to.something.else",
				isReplaced: true,
				prefixSrc:  "prefix1.from.",
				prefixDst:  "prefix2.to.",
			}},
		},

		// replace entire string
		{
			config: map[string]string{
				"prefix1.from.": "prefix2.to.",
				"aaa.":          "111.",
				"bbb.":          "222.",
			},
			pattern: "prefix1.from.",
			expected: []SubstituteInfo{{
				MetricSrc:  "prefix1.from.",
				MetricDst:  "prefix2.to.",
				isReplaced: true,
				prefixSrc:  "prefix1.from.",
				prefixDst:  "prefix2.to.",
			}},
		},

		// truncate prefix
		{
			config: map[string]string{
				"prefix1.from.": "",
				"aaa.":          "111.",
				"bbb.":          "222.",
			},
			pattern: "prefix1.from.something.else",
			expected: []SubstituteInfo{{
				MetricSrc:  "prefix1.from.something.else",
				MetricDst:  "something.else",
				isReplaced: true,
				prefixSrc:  "prefix1.from.",
				prefixDst:  "",
			}},
		},

		// truncate entire string
		{
			config: map[string]string{
				"prefix1.from.": "",
				"aaa.":          "111",
				"bbb.":          "222",
			},
			pattern: "prefix1.from.",
			expected: []SubstituteInfo{{
				MetricSrc:  "prefix1.from.",
				MetricDst:  "",
				isReplaced: true,
				prefixSrc:  "prefix1.from.",
				prefixDst:  "",
			}},
		},

		//
		// replacement with pattern prefix matching
		//

		// replace prefix
		{
			config: map[string]string{
				"prefix.from1.": "prefix.to1.",
				"prefix.from2.": "prefix.to2.",
				"prefix.from3.": "prefix.to3.",
			},
			pattern: "prefix.*.something.else",
			expected: []SubstituteInfo{
				{
					MetricSrc:  "prefix.from1.something.else",
					MetricDst:  "prefix.to1.something.else",
					isReplaced: true,
					prefixSrc:  "prefix.from1.",
					prefixDst:  "prefix.to1.",
				},
				{
					MetricSrc:  "prefix.from2.something.else",
					MetricDst:  "prefix.to2.something.else",
					isReplaced: true,
					prefixSrc:  "prefix.from2.",
					prefixDst:  "prefix.to2.",
				},
				{
					MetricSrc:  "prefix.from3.something.else",
					MetricDst:  "prefix.to3.something.else",
					isReplaced: true,
					prefixSrc:  "prefix.from3.",
					prefixDst:  "prefix.to3.",
				},
			},
		},

		// replace entire string
		{
			config: map[string]string{
				"a.b.c.d.from1.": "a.b.c.d.to1.",
				"a.b.c.d.from2.": "a.b.c.d.to2.",
				"a.b.c.d.from3.": "a.b.c.d.to3.",
			},
			pattern: "a.b.c.d.*.",
			expected: []SubstituteInfo{
				{
					MetricSrc:  "a.b.c.d.from1.",
					MetricDst:  "a.b.c.d.to1.",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from1.",
					prefixDst:  "a.b.c.d.to1.",
				},
				{
					MetricSrc:  "a.b.c.d.from2.",
					MetricDst:  "a.b.c.d.to2.",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from2.",
					prefixDst:  "a.b.c.d.to2.",
				},
				{
					MetricSrc:  "a.b.c.d.from3.",
					MetricDst:  "a.b.c.d.to3.",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from3.",
					prefixDst:  "a.b.c.d.to3.",
				},
			},
		},

		// truncate prefix
		{
			config: map[string]string{
				"a.b.c.d.from1.": "a.b.c.d.to1.",
				"a.b.c.d.from2.": "",
			},
			pattern: "a.b.c.d.*.something.else",
			expected: []SubstituteInfo{
				{
					MetricSrc:  "a.b.c.d.from1.something.else",
					MetricDst:  "a.b.c.d.to1.something.else",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from1.",
					prefixDst:  "a.b.c.d.to1.",
				},
				{
					MetricSrc:  "a.b.c.d.from2.something.else",
					MetricDst:  "something.else",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from2.",
					prefixDst:  "",
				},
			},
		},

		// truncate entire string
		{
			config: map[string]string{
				"a.b.c.d.from1.": "a.b.c.d.to1.",
				"a.b.c.d.from2.": "",
			},
			pattern: "a.b.c.d.*.",
			expected: []SubstituteInfo{
				{
					MetricSrc:  "a.b.c.d.from1.",
					MetricDst:  "a.b.c.d.to1.",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from1.",
					prefixDst:  "a.b.c.d.to1.",
				},
				{
					MetricSrc:  "a.b.c.d.from2.",
					MetricDst:  "",
					isReplaced: true,
					prefixSrc:  "a.b.c.d.from2.",
					prefixDst:  "",
				},
			},
		},
	}

	for i, testCase := range testCases {
		patternProcessor := NewPatternProcessor(testCase.config)
		substitutes := patternProcessor.ReplacePrefix(testCase.pattern)

		// assert equality disregarding elements order
		assert.ElementsMatch(t, testCase.expected, substitutes, fmt.Sprintf("Test case #%d", i+1))
	}
}

func TestPatternProcessor_RestoreMetricNameSimplePattern(t *testing.T) {
	type TestCase struct {
		config   map[string]string
		pattern  string
		expected []string
	}

	testCases := []TestCase{
		// no replacement
		{
			config: map[string]string{
				"prefix.from.":  "prefix.to.",
				"prefix1.from.": "prefix2.to.",
			},
			pattern: "a.b.c.d",
			expected: []string{
				"a.b.c.d",
			},
		},

		// replacement with strict prefix matching
		{
			config: map[string]string{
				"prefix1.from.": "prefix2.to.",
				"aaa.":          "111.",
				"bbb.":          "222.",
			},
			pattern: "prefix1.from.something.else",
			expected: []string{
				"prefix1.from.something.else",
			},
		},
		{
			config: map[string]string{
				"prefix1.from.": "prefix2.to.",
			},
			pattern: "prefix1.from.",
			expected: []string{
				"prefix1.from.",
			},
		},
		{
			config: map[string]string{
				"prefix1.from.": "",
			},
			pattern: "prefix1.from.something.else",
			expected: []string{
				"prefix1.from.something.else",
			},
		},
		{
			config: map[string]string{
				"prefix1.from.": "",
			},
			pattern: "prefix1.from.",
			expected: []string{
				"prefix1.from.",
			},
		},

		// replacement with pattern prefix matching
		{
			config: map[string]string{
				"prefix.from1.": "prefix.to1.",
				"prefix.from2.": "prefix.to2.",
				"prefix.from3.": "prefix.to3.",
			},
			pattern: "prefix.*.something.else",
			expected: []string{
				"prefix.from1.something.else",
				"prefix.from2.something.else",
				"prefix.from3.something.else",
			},
		},
		{
			config: map[string]string{
				"a.b.c.d.from1.": "a.b.c.d.to1.",
				"a.b.c.d.from2.": "a.b.c.d.to2.",
				"a.b.c.d.from3.": "a.b.c.d.to3.",
			},
			pattern: "a.b.c.d.*.",
			expected: []string{
				"a.b.c.d.from1.",
				"a.b.c.d.from2.",
				"a.b.c.d.from3.",
			},
		},
		{
			config: map[string]string{
				"a.b.c.d.from1.": "a.b.c.d.to1.",
				"a.b.c.d.from2.": "",
			},
			pattern: "a.b.c.d.*.something.else",
			expected: []string{
				"a.b.c.d.from1.something.else",
				"a.b.c.d.from2.something.else",
			},
		},
		{
			config: map[string]string{
				"a.b.c.d.from1.": "a.b.c.d.to1.",
				"a.b.c.d.from2.": "",
			},
			pattern: "a.b.c.d.*.",
			expected: []string{
				"a.b.c.d.from1.",
				"a.b.c.d.from2.",
			},
		},
	}

	for i, testCase := range testCases {
		patternProcessor := NewPatternProcessor(testCase.config)
		substitutes := patternProcessor.ReplacePrefix(testCase.pattern)

		restoredMetrics := make([]string, 0, len(substitutes))
		for _, substitute := range substitutes {
			restoredMetrics = append(restoredMetrics, patternProcessor.RestoreMetricName(substitute.MetricDst, substitute))
		}

		// assert equality disregarding elements order
		assert.ElementsMatch(t, testCase.expected, restoredMetrics, fmt.Sprintf("Test case #%d", i+1))
	}
}
