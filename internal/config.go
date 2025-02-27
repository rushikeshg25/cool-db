package internal

import "sync"

type CoolConfig struct {
	DbFile  string
	Version string
	Wg      sync.WaitGroup
}
