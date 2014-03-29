// +build !linux,!freebsd !cgo

package graphdb

func NewSqliteConn(root string) (*Database, error) {
	panic("Not implemented")
}
