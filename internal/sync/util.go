package sync

import (
	"fmt"

	"github.com/ImTheCurse/ConflowCI/internal/provider/pb"
)

func GetProtoWorkerError(prefix string, err error, resp *pb.SyncResponse) string {
	if resp.Error != nil {
		return fmt.Sprintf("%s: Worket build error: %s", prefix, resp.Error.Reason)
	}
	if err != nil {
		return fmt.Sprintf("%s: Worker build error: proto server error: %v", prefix, err)
	}
	return ""
}
