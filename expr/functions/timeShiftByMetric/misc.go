package timeShiftByMetric

import (
	"sort"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/types"
)

type callParams struct {
	marks       []*types.MetricData
	metrics     []*types.MetricData
	versionRank int
	pointsQty   int
	stepTime    int32
}

type versionInfo struct {
	position     int
	versionMajor int
	versionMinor int
}

type versionInfos []versionInfo

// HighestVersions returns slice of markVersionInfo containing the highest version
// for each major version with limit (-1 means no limit)
func (data versionInfos) HighestVersions(limit int) versionInfos {
	qty := 0
	result := make(versionInfos, 0, len(data))

	sort.Sort(sort.Reverse(data))
	for i := 0; (i < len(data)) && (qty < limit || limit == -1); i++ {
		if (qty == 0) || (result[qty-1].versionMajor != data[i].versionMajor) {
			result = append(result, data[i])
			qty++
		}
	}

	return result
}

func (data versionInfos) Len() int {
	return len(data)
}

func (data versionInfos) Less(i, j int) bool {
	if data[i].versionMajor == data[j].versionMajor {
		return data[i].versionMinor < data[j].versionMinor
	} else {
		return data[i].versionMajor < data[j].versionMajor
	}
}

func (data versionInfos) Swap(i, j int) {
	data[i], data[j] = data[j], data[i]
}

// mustAtoi is like strconv.Atoi, but causes panic in case of error
func mustAtoi(s string) int {
	if result, err := strconv.Atoi(s); err != nil {
		panic(err)
	} else {
		return result
	}
}
