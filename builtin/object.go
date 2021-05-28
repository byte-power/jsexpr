package builtin

import "time"

var objects = map[string]interface{}{
	"Date": dateObject{
		Now: time.Now().Unix,
	},
}

func Objs() map[string]interface{} {
	return objects
}

type dateObject struct {
	Now func() int64 `jsexpr:"now"`
}

// func (this dateObject) Now() int64 {
// 	return time.Now().Unix()
// }
