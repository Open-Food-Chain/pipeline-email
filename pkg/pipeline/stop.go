package pipeline

func (p *Pipeline) Stop() {
	p.StopChannel <- true
}
