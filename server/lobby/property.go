package lobby

import (
	"bytes"

	"wsnet2/binary"
)

type PropQuery struct {
	Key string
	Op  string
	Val []byte
}

func (q *PropQuery) test(val []byte) bool {
	if binary.Type(q.Val[0]) != binary.Type(val[0]) {
		return false
	}
	ret := bytes.Compare(val[1:], q.Val[1:])
	switch q.Op {
	case "=":
		return ret == 0
	case "!":
		return ret != 0
	case "<":
		return ret < 0
	case "<=":
		return ret <= 0
	case ">":
		return ret > 0
	case ">=":
		return ret >= 0
	}
	return false
}

type PropQueries []PropQuery

func (pqs *PropQueries) test(props binary.Dict) bool {
	for _, q := range *pqs {
		if !q.test(props[q.Key]) {
			return false
		}
	}
	return true
}
