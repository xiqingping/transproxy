package transproxy

import (
	"net"
	"sync"
)

func sliceEqual(a, b []byte) bool {
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

// BlackList black list of IP
type BlackList struct {
	black []net.IP
	lock  *sync.RWMutex
}

// NewBlackList create a black list
func NewBlackList() *BlackList {
	return &BlackList{
		black: make([]net.IP, 0, 500),
		lock:  new(sync.RWMutex),
	}
}

// Contains check if black list contains the specified IP
func (bl *BlackList) Contains(ip net.IP) bool {
	bl.lock.RLock()
	defer bl.lock.RUnlock()

	for _, v := range bl.black {
		if sliceEqual(v, ip) {
			return true
		}
	}

	return false
}

// Add add a IP to the black list
func (bl *BlackList) Add(ip net.IP) {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	for _, v := range bl.black {
		if sliceEqual(v, ip) {
			return
		}
	}

	bl.black = append(bl.black, ip)
}
