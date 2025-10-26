package sync

import (
	"fmt"
	"sync"

	"github.com/ImTheCurse/ConflowCI/internal/provider/pb"
)

func GetProtoWorkerError(prefix string, err error, resp *pb.SyncResponse) string {
	if resp == nil {
		return fmt.Sprintf("%s: Worker build error: %v", prefix, err)
	}
	if resp.Error != nil {
		return fmt.Sprintf("%s: Worket build error: %s", prefix, resp.Error.Reason)
	}
	if err != nil {
		return fmt.Sprintf("%s: Worker build error: proto server error: %v", prefix, err)
	}
	return ""
}

func ConcurrentAppendToArray[T any](mu *sync.Mutex, val T, arr *[]T) {
	mu.Lock()
	*arr = append(*arr, val)
	mu.Unlock()
}
