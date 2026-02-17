package convert

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// PbToMap: convert a protobuf Message to map
// if opt == nil, use default MarshalOptions:
//
//	&protojson.MarshalOptions{
//		UseProtoNames: true,
//	}
func PbToMap[T any](msg proto.Message, opt *protojson.MarshalOptions) (map[string]T, error) {
	if opt == nil {
		opt = &protojson.MarshalOptions{
			UseProtoNames: true,
		}
	}

	data, err := opt.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("protojson marshal: %w", err)
	}

	result := make(map[string]T)
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	return result, nil
}

func MustPbCopy[S any, D any](src *S) *D {
	if src == nil {
		return nil
	}

	dst := new(D)
	err := copier.CopyWithOption(dst, src, copier.Option{
		DeepCopy: true,
		Converters: []copier.TypeConverter{
			pbToTime,
			timeToPb,
			durationToPb,
			pbToDuration,
			gormDeletedAtToPb,
		},
	})
	if err != nil {
		panic(err)
	}
	return dst
}
