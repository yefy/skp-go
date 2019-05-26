package mq

type Vector struct {
	buffer []byte
}

func NewVector() *Vector {
	v := &Vector{}
	v.buffer = make([]byte, 0, 4096)
	return v
}

func (v *Vector) Put(buf []byte) {
	v.buffer = append(v.buffer, buf...)
}

func (v *Vector) Get(size int) []byte {
	buffLen := len(v.buffer)
	if size > buffLen {
		size = buffLen
	}
	return v.buffer[:size]
}

func (v *Vector) GetAll() []byte {
	buffLen := len(v.buffer)
	return v.buffer[:buffLen]
}

func (v *Vector) Skip(size int) {
	buffLen := len(v.buffer)
	if size > buffLen {
		size = buffLen
	}
	v.buffer = v.buffer[size:]
}

func (v *Vector) SkipAll() {
	buffLen := len(v.buffer)
	v.buffer = v.buffer[buffLen:]
}
