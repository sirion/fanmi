package config

type Entry struct {
	Temp  float32
	Speed float32
}

type Values []Entry

func (v Values) Len() int {
	return len(v)
}

func (v Values) Less(a int, b int) bool {
	return v[a].Temp < v[b].Temp
}

func (v Values) Swap(a int, b int) {
	v[a], v[b] = v[b], v[a]
}
