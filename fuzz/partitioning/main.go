package partitioning

import tfw "github.com/tsnow/typing-for-war/cmd/tfw"

func Fuzz(data []byte) int {
	dlen := len(data) / 2
	tfw.GoodBadLeft(string(data[:dlen]),string(data[dlen:]))
	return 0
}
