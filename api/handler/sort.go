package handler

import "github.com/winston-ci/redgreen/api/builds"

type ByCreatedAt []builds.Build

func (builds ByCreatedAt) Len() int {
	return len(builds)
}

func (builds ByCreatedAt) Less(i, j int) bool {
	return builds[i].CreatedAt.UnixNano() < builds[j].CreatedAt.UnixNano()
}

func (builds ByCreatedAt) Swap(i, j int) {
	builds[i], builds[j] = builds[j], builds[i]
}
