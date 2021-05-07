package common

type Event struct {
	Market     string  `json:"m" msgpack:"m"`
	Instrument string  `json:"i" msgpack:"i"`
	Timestamp  uint64  `json:"t" msgpack:"t"`
	Action     Action  `json:"a" msgpack:"a"`
	Units      float64 `json:"u" msgpack:"u"`
	Price      float64 `json:"u" msgpack:"p"`
}
