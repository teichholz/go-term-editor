package rope

type RopeWriter struct {
	rope *Rope
}

func Writer() *RopeWriter {
	rope := New()
	return &RopeWriter{rope: &rope}
}

func (writer *RopeWriter) Write(p []byte) (n int, err error) {
	rp := writer.rope.AppendString(string(p))
	writer.rope = &rp
	return len(p), nil
}

func (writer *RopeWriter) Rope() *Rope {
	return writer.rope
}
