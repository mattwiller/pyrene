package system

import "strconv"

type Integer int64

var _ Value = Integer(42)

func (Integer) Type() ValueType {
	return IntegerType
}

func (n Integer) String() string {
	return strconv.FormatInt(int64(n), 10)
}
