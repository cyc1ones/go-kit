package convert

// definitions of copier.TypeConverter

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// time.Time
var pbToTime = copier.TypeConverter{
	SrcType: &timestamppb.Timestamp{},
	DstType: time.Time{},

	Fn: func(src interface{}) (dst interface{}, err error) {
		return src.(*timestamppb.Timestamp).AsTime(), nil
	},
}

var timeToPb = copier.TypeConverter{
	SrcType: time.Time{},
	DstType: &timestamppb.Timestamp{},

	Fn: func(src interface{}) (dst interface{}, err error) {
		return timestamppb.New(src.(time.Time)), nil
	},
}

// gorm.DeletedAt
var gormDeletedAtToPb = copier.TypeConverter{
	SrcType: gorm.DeletedAt{},
	DstType: &timestamppb.Timestamp{},
	Fn: func(src interface{}) (dst interface{}, err error) {
		da, ok := src.(gorm.DeletedAt)
		if !ok {
			return nil, fmt.Errorf("src is not gorm.DeletedAt")
		}

		if da.Valid {
			return timestamppb.New(da.Time), nil
		}
		return nil, nil
	},
}

// time.Duration
var durationToPb = copier.TypeConverter{
	SrcType: time.Duration(0),
	DstType: &durationpb.Duration{},
	Fn: func(src interface{}) (dst interface{}, err error) {
		return durationpb.New(src.(time.Duration)), nil
	},
}

var pbToDuration = copier.TypeConverter{
	SrcType: new(durationpb.Duration),
	DstType: time.Duration(0),

	Fn: func(src interface{}) (dst interface{}, err error) {
		return src.(*durationpb.Duration).AsDuration(), nil
	},
}
