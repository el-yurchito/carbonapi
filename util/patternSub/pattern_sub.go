package patternSub

import (
	"fmt"
	"strings"
)

const (
	pathSep  = "."
	tagsSep  = ";"
	wildcard = "*"

	sbtStart = "seriesByTag("
	sbtEnd   = ")"
	sbtSep   = "','"

	argsSep = ","
	quote   = "'"
)

var (
	// available signs for `seriesByTag` argument terms
	argsSigns = map[string]bool{
		"=":   true,
		"!=":  true,
		"=~":  true,
		"!=~": true,
	}

	// available options for `name` tag
	nameTagOptions = map[string]bool{
		"name":     true,
		"__name__": true,
	}
)

type nameTagInfo struct {
	index int
	sign  string
}

type subMap map[string]string

// PatternProcessor is used to replace prefixes of time series patterns
// e.g. prefix1.a.b.c. -> prefix2.d.e.f.
// all replacements must end with period
// although target replacement can be empty (but source can not)
type PatternProcessor struct {
	config subMap            // original substitute configuration
	prefix map[string]subMap // config with substitute keys grouped by common prefixes
}

// SubstituteInfo keeps information about one metric replacement
type SubstituteInfo struct {
	MetricSrc  string
	MetricDst  string
	isReplaced bool
	prefixSrc  string
	prefixDst  string
	tagInfo    *nameTagInfo
}

// NewPatternProcessor makes new PatternProcessor based on replacement config
func NewPatternProcessor(config map[string]string) *PatternProcessor {
	qty := len(config)
	result := &PatternProcessor{
		config: make(subMap, qty),
		prefix: make(map[string]subMap, qty+1),
	}

	for replaceFrom, replaceTo := range config {
		// sanity check: both replaceFrom and replaceTo should end with period
		// yet replaceTo can also be empty
		if !(strings.HasSuffix(replaceFrom, pathSep) && (strings.HasSuffix(replaceTo, pathSep) || replaceTo == "")) {
			continue
		}

		result.config[replaceFrom] = replaceTo // copy original config

		replacePrefix := result.cutOffLastNode(replaceFrom)
		if replacePrefix == "" { // no extra processing if pattern consists of exactly one node
			continue
		}

		replacePrefix = replacePrefix + pathSep + wildcard + pathSep
		if _, exists := result.prefix[replacePrefix]; !exists {
			result.prefix[replacePrefix] = make(subMap, qty)
		}

		result.prefix[replacePrefix][replaceFrom] = replaceTo
	}

	// default group
	result.prefix[""] = result.config

	// leave only groups containing more than one rewrite options
	// groups with no more than one rewrite options are dispatched to default
	for prefix, replaceMap := range result.prefix {
		if prefix != "" && len(replaceMap) <= 1 {
			delete(result.prefix, prefix)
		}
	}

	return result
}

func (pp *PatternProcessor) GetDefaultSubstituteMap() map[string]string {
	return pp.prefix[""]
}

// ReplacePrefix determines whether pattern matches any group prefix
// (e.g. pattern = a.b.*.c.d and prefix is a.b.*)
// if group prefix is matched then function generate substitutes based on each key in group
// if prefix isn't matched then function attempts to use default substitute
func (pp *PatternProcessor) ReplacePrefix(pattern string) []SubstituteInfo {
	if strings.HasPrefix(pattern, sbtStart) && strings.HasSuffix(pattern, sbtEnd) {
		return pp.replacePrefixFunctionArg(pattern)
	} else {
		return pp.replacePrefixSimplePattern(pattern)
	}
}

// RestoreMetricName rolls back prefix substitute of metric name
func (pp *PatternProcessor) RestoreMetricName(metricName string, substituteInfo SubstituteInfo) string {
	if !substituteInfo.isReplaced { // there wasn't any substitute - no need to roll anything back
		return metricName
	}

	// simple pattern
	if substituteInfo.tagInfo == nil {
		return pp.restoreMetricNameSimplePattern(metricName, substituteInfo)
	}

	// function call looks like `seriesByTag('name=a.b.c.d', 'tag1=val1', 'tag2=val2')`
	// but resulting metrics look like `a.b.c.d;tag1=val1;tag2=val2`
	tagParts := strings.SplitN(pp.cleanFunctionCall(metricName), tagsSep, 1)
	if len(tagParts) == 1 {
		return pp.restoreMetricNameSimplePattern(tagParts[0], substituteInfo)
	}

	return metricName
}

// cleanArg removes possible leading and trailing quotes
func (pp *PatternProcessor) cleanArg(arg string) string {
	return strings.TrimPrefix(strings.TrimSuffix(arg, quote), quote)
}

// cleanFunctionCall gets rid of useless parts of function call
func (pp *PatternProcessor) cleanFunctionCall(pattern string) string {
	pattern = strings.TrimPrefix(pattern, sbtStart)
	pattern = strings.TrimSuffix(pattern, sbtEnd)
	pattern = strings.ReplaceAll(pattern, " ", "")
	return pattern
}

// cutOffLastNode takes metric path (e.g. "a.b.c.d."), cuts off its last node and returns the rest
func (pp *PatternProcessor) cutOffLastNode(metricPath string) string {
	index := strings.LastIndexByte(strings.TrimSuffix(metricPath, pathSep), pathSep[0])
	if index == -1 {
		return metricPath
	} else {
		return metricPath[:index]
	}
}

// replacePrefixFunctionArg is version of ReplacePrefix which works only for `seriesByTag` function call
// changes are applied only to `name` tag if there is one
func (pp *PatternProcessor) replacePrefixFunctionArg(pattern string) []SubstituteInfo {
	var (
		args []string

		// attributes of located `name` tag (if any)
		tagInfo           *nameTagInfo
		tagName, tagValue string
	)

	pattern = pp.cleanFunctionCall(pattern)
	args = strings.Split(pattern, sbtSep)

	for i, arg := range args {
		arg = pp.cleanArg(arg)
		name, value, sign, err := pp.splitArgTerm(arg)
		if err != nil {
			continue
		}

		if tagInfo == nil && nameTagOptions[name] { // process only the first `name` tag if there are more than one
			tagInfo = &nameTagInfo{
				index: i,
				sign:  sign,
			}
			tagName = name
			tagValue = value
		} else {
			args[i] = quote + name + sign + value + quote
		}
	}

	if tagInfo != nil {
		simpleReplacementList := pp.replacePrefixSimplePattern(tagValue) // do simple replacement for parsed value
		result := make([]SubstituteInfo, 0, len(simpleReplacementList))

		// all parts except related to `name` tag will be the same
		// so it's possible to distinguish common prefix and suffix
		commonPrefix := sbtStart + strings.Join(args[:tagInfo.index], argsSep)
		commonSuffix := strings.Join(args[tagInfo.index+1:], argsSep) + sbtEnd

		for _, simpleReplacement := range simpleReplacementList {
			metricSrc := commonPrefix
			metricDst := commonPrefix
			if tagInfo.index > 0 { // sep isn't placed if `name` if the first argument
				metricSrc += argsSep
				metricDst += argsSep
			}

			metricSrc += quote + tagName + tagInfo.sign + simpleReplacement.MetricSrc + quote
			metricDst += quote + tagName + tagInfo.sign + simpleReplacement.MetricDst + quote
			if tagInfo.index < len(args)-1 { // sep isn't place if `name` is the last argument
				metricSrc += argsSep
				metricDst += argsSep
			}

			metricSrc += commonSuffix
			metricDst += commonSuffix

			result = append(result, SubstituteInfo{
				MetricSrc:  metricSrc,
				MetricDst:  metricDst,
				isReplaced: simpleReplacement.isReplaced,
				prefixSrc:  simpleReplacement.prefixSrc,
				prefixDst:  simpleReplacement.prefixDst,
				tagInfo:    tagInfo,
			})
		}

		return result
	} else {
		pattern = sbtStart + strings.Join(args, argsSep) + sbtEnd
		return []SubstituteInfo{{
			MetricSrc:  pattern,
			MetricDst:  pattern,
			isReplaced: false,
			prefixSrc:  "",
			prefixDst:  "",
			tagInfo:    nil,
		}}
	}
}

// replacePrefixSimplePattern is version of ReplacePrefix which works only with parsed patterns
// i.e. those ones which are not inside any function call
func (pp *PatternProcessor) replacePrefixSimplePattern(pattern string) []SubstituteInfo {
	var (
		matchedReplaceMap subMap
		matchedPrefix     string
		result            []SubstituteInfo
	)

	// search for replacement group which has the same prefix the pattern does
	for prefix, replaceMap := range pp.prefix {
		if prefix != "" && strings.HasPrefix(pattern, prefix) {
			matchedReplaceMap = replaceMap
			matchedPrefix = prefix
			break
		}
	}

	result = make([]SubstituteInfo, 0, len(matchedReplaceMap)+1)
	if matchedReplaceMap != nil { // found grouping prefix, will produce substitutes based on each key in group
		patternSuffix := strings.TrimPrefix(pattern, matchedPrefix) // the suffix is common for all substitutes
		for replaceFrom, replaceTo := range matchedReplaceMap {
			result = append(result, SubstituteInfo{
				MetricSrc:  replaceFrom + patternSuffix,
				MetricDst:  replaceTo + patternSuffix,
				isReplaced: true,
				prefixSrc:  replaceFrom,
				prefixDst:  replaceTo,
			})
		}
	} else { // use default substitute
		var (
			defaultReplaceMap subMap
			substituteInfo    *SubstituteInfo
		)

		defaultReplaceMap = pp.prefix[""]
		for replaceFrom, replaceTo := range defaultReplaceMap {
			if strings.HasPrefix(pattern, replaceFrom) {
				substituteInfo = &SubstituteInfo{
					MetricSrc:  pattern,
					MetricDst:  replaceTo + strings.TrimPrefix(pattern, replaceFrom),
					isReplaced: true,
					prefixSrc:  replaceFrom,
					prefixDst:  replaceTo,
				}
				break
			}
		}

		if substituteInfo == nil {
			substituteInfo = &SubstituteInfo{
				MetricSrc:  pattern,
				MetricDst:  pattern,
				isReplaced: false,
				prefixSrc:  "",
				prefixDst:  "",
			}
		}
		result = append(result, *substituteInfo)
	}

	return result
}

// restoreMetricNameSimplePattern restores simple replacement of metric name
func (pp *PatternProcessor) restoreMetricNameSimplePattern(metricName string, substituteInfo SubstituteInfo) string {
	return substituteInfo.prefixSrc + strings.TrimPrefix(metricName, substituteInfo.prefixDst)
}

// splitArgTerm splits term like 'tag=value' and returns its parts
func (pp *PatternProcessor) splitArgTerm(term string) (tagName string, tagValue string, sign string, err error) {
	sign = "="

	parts := strings.SplitN(term, sign, 2)
	if len(parts) != 2 {
		err = fmt.Errorf("bad argument format: %#v (invalid number of parts)", term)
		return
	}

	if len(parts[0]) > 0 && parts[0][len(parts[0])-1] == '!' { // optional `!` (`!=` or `!=~`)
		sign = "!" + sign
		parts[0] = parts[0][:len(parts[0])-1]
	}

	if len(parts[1]) > 0 && parts[1][0] == '~' { // optional `~` (`=~` or `!=~`)
		sign = sign + "~"
		parts[1] = parts[1][1:]
	}

	if !argsSigns[sign] {
		err = fmt.Errorf("bad argument format: %#v (invalid sign)", term)
		return
	}

	tagName = parts[0]
	tagValue = parts[1]
	return
}
