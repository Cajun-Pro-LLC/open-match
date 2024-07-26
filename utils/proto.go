package utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Int32ToProto(v int32) []byte {
	value := &wrapperspb.Int32Value{
		Value: v,
	}
	b, err := proto.Marshal(value)
	if err != nil {
		log.Err(err).Int32("value", v).Msg("Failed to encode to proto")
		return []byte{}
	}
	return b
}

func ProtoToInt32(v *anypb.Any, fallback int32) int32 {
	if v == nil {
		return fallback
	}
	var value wrapperspb.Int32Value
	err := proto.Unmarshal(v.GetValue(), &value)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal")
		return fallback
	}
	return value.GetValue()
}

func BoolToProto(v bool) []byte {
	value := &wrapperspb.BoolValue{
		Value: v,
	}
	b, err := proto.Marshal(value)
	if err != nil {
		log.Err(err).Bool("value", v).Msg("Failed to encode to proto")
		return []byte{}
	}
	return b
}

func ProtoToBool(v *anypb.Any, fallback bool) bool {
	if v == nil {
		return fallback
	}
	var value wrapperspb.BoolValue
	err := proto.Unmarshal(v.GetValue(), &value)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal")
		return fallback
	}
	return value.GetValue()
}
