package vm

type PropertyProvider interface {
	FetchProperty(property string) interface{}
}

type ValueProvider interface {
	GetValue() interface{}
}
