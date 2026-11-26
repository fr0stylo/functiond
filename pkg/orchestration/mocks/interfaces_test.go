package mocks_test

import (
	"functiond/pkg/orchestration"
	"functiond/pkg/orchestration/mocks"
)

var (
	_ orchestration.Executor   = (*mocks.MockExecutor)(nil)
	_ orchestration.WorkersSet = (*mocks.MockWorkersSet)(nil)
)
