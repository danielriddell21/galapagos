package nn

func (m *MLP) TrainBatch(opt *Adam, xs [][]float64, targets []int) (loss float64) {
	g := m.newGrads()
	for b := range xs {
		out, acts, zs := m.forwardCached(xs[b])
		l, dOut := SoftmaxCrossEntropy(out, targets[b])
		loss += l
		m.backwardInto(g, acts, zs, dOut)
	}
	n := float64(len(xs))
	g.scale(n)
	opt.Step(m, g)
	return loss / n
}
