package groupLeftTagged

import (
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
)

const nameTag = "name"

type taggedMetric struct {
	tags  map[string]string
	order []string
}

func parseTaggedMetric(metric string) *taggedMetric {
	metric = helper.ExtractMetric(metric)
	metricRunes := []rune(metric)
	result := &taggedMetric{
		tags:  make(map[string]string),
		order: make([]string, 0, 4),
	}

	isTagValue := true
	name := nameTag
	token := strings.Builder{}
	for _, char := range metricRunes {
		if char == ';' { // new name=value pair starts, save the current one
			result.updateOne(name, token.String())
			token.Reset()
			isTagValue = false
		} else if char == '=' { // tag's name changes to tag's value
			name = token.String()
			token.Reset()
			isTagValue = true
		} else {
			token.WriteRune(char)
		}
	}

	// the last token left
	if isTagValue {
		result.updateOne(name, token.String())
	}

	return result
}

func (tm *taggedMetric) String() string {
	result := strings.Builder{}
	result.WriteString(tm.tags[nameTag])

	for _, tag := range tm.order {
		if tag == nameTag {
			continue
		}

		result.WriteByte(';')
		result.WriteString(tag)
		result.WriteByte('=')
		result.WriteString(tm.tags[tag])
	}

	return result.String()
}

func (tm *taggedMetric) clone() *taggedMetric {
	result := &taggedMetric{
		tags:  make(map[string]string, len(tm.tags)),
		order: make([]string, 0, len(tm.order)),
	}
	for key, val := range tm.tags {
		result.tags[key] = val
	}
	for _, tag := range tm.order {
		result.order = append(result.order, tag)
	}
	return result
}

func (tm *taggedMetric) key(tags []string) string {
	tagsQty := len(tags)
	result := strings.Builder{}
	for i, tag := range tags {
		result.WriteString(tm.tags[tag])
		if i != tagsQty-1 {
			result.WriteByte('.')
		}
	}
	return result.String()
}

func (tm *taggedMetric) merge(another *taggedMetric) *taggedMetric {
	result := tm.clone()
	for _, tag := range another.order {
		if tag != nameTag {
			result.updateOne(tag, another.tags[tag])
		}
	}
	return result
}

func (tm *taggedMetric) updateOne(key, value string) {
	if _, ok := tm.tags[key]; !ok {
		tm.order = append(tm.order, key)
	}
	tm.tags[key] = value
}
