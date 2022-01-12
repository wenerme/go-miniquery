package miniquery

func Parse(s string) (*Node, error) {
	p := &MiniQueryPeg{Tree: &Tree{}, Buffer: s}
	if err := p.Init(); err != nil {
		return nil, err
	}
	if err := p.Parse(); err != nil {
		return nil, err
	}
	p.Execute()
	if len(p.Errors) != 0 {
		return nil, p.Errors[0]
	}
	return p.Pop(), nil
}
