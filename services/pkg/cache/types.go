package cache

type PostRecord struct {
	AtURI     string
	Timestamp int64 `msgpack:"t"`
}

func (p PostRecord) IsEmpty() bool {
	return p.AtURI == "" || p.Timestamp == 0
}
