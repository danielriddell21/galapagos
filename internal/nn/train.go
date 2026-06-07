package nn

// TrainBatch runs one optimizer step on a minibatch of classification examples
// (softmax cross-entropy) and returns the mean loss. xs are inputs, targets are
// class indices; both must have the same length.
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
