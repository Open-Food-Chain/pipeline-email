package pipeline

func (p *Pipeline) Stop() {
	p.stopChannel <- true
}
