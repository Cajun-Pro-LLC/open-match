package utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Int32ToProto(v int32) []byte {
	value := &wrapperspb.Int32Value{
		Value: v,
	}
	b, err := proto.Marshal(value)
	if err != nil {
		log.Err(err).Int32("value", v).Msg("Failed to encode to proto")
	}
	return b
}
