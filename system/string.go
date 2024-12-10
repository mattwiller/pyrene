package system

type String string

var _ Value = String("")

func (String) Type() ValueType {
	return StringType
}

func (s String) String() string {
	return string(`'` + s + `'`)
}
