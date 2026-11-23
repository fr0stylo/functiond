package orchestration

import "sync"

type Set struct {
	m sync.Map
}
