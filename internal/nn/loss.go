package nn

import "math"

// Softmax returns the softmax of logits (numerically stable).
func Softmax(logits []float64) []float64 {
	mx := logits[0]
	for _, v := range logits {
		mx = max(mx, v)
	}
	out := make([]float64, len(logits))
	var sum float64
	for i, v := range logits {
		e := math.Exp(v - mx)
		out[i] = e
		sum += e
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

// SoftmaxCrossEntropy returns the cross-entropy loss of logits against the
// target class, and the gradient of the loss w.r.t. the logits, which is the
// clean softmax(logits) − onehot(target).
func SoftmaxCrossEntropy(logits []float64, target int) (loss float64, dLogits []float64) {
	p := Softmax(logits)
	loss = -math.Log(max(p[target], 1e-12))
	dLogits = p // reuse
	dLogits[target]--
	return loss, dLogits
}

// MSE returns the mean squared error and its gradient w.r.t. pred.
func MSE(pred, target []float64) (loss float64, dPred []float64) {
	n := float64(len(pred))
	dPred = make([]float64, len(pred))
	for i := range pred {
		d := pred[i] - target[i]
		loss += d * d
		dPred[i] = 2 * d / n
	}
	return loss / n, dPred
}
