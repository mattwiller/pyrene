package system

type Boolean bool

var _ Value = Boolean(true)

func (Boolean) Type() ValueType {
	return BooleanType
}

func (b Boolean) String() string {
	if b {
		return "true"
	} else {
		return "false"
	}
}
