package util

import (
	"sort"

	pb "github.com/Mojjedrengen/ChitChat/grpc"
)

func SortMsgListBasedOnTime(list []*pb.Msg) []*pb.Msg {
	sort.Slice(list, func(i, j int) bool {
		return list[i].UnixTime < list[j].UnixTime
	})
	return list
}
