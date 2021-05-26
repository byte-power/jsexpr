package file

type Location struct {
	Line   int `msgpack:"line"`   // The 1-based line of the location.
	Column int `msgpack:"column"` // The 0-based column number of the location.
}

func (l Location) Empty() bool {
	return l.Column == 0 && l.Line == 0
}
