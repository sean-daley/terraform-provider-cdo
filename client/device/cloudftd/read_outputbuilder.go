package cloudftd

type ReadOutputBuilder struct {
	readOutput *ReadOutput
}

func NewReadOutputBuilder() *ReadOutputBuilder {
	readOutput := &ReadOutput{}
	b := &ReadOutputBuilder{readOutput: readOutput}
	return b
}

func (b *ReadOutputBuilder) Uid(uid string) *ReadOutputBuilder {
	b.readOutput.Uid = uid
	return b
}

func (b *ReadOutputBuilder) Name(name string) *ReadOutputBuilder {
	b.readOutput.Name = name
	return b
}

func (b *ReadOutputBuilder) Metadata(metadata Metadata) *ReadOutputBuilder {
	b.readOutput.Metadata = metadata
	return b
}

func (b *ReadOutputBuilder) Build() ReadOutput {
	return *b.readOutput
}
