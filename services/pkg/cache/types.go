package cache

type PostRecord struct {
	AtURI     string
	Timestamp int `msgpack:"t"`
}
