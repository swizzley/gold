package layers

import (
	"github.com/pbarker/go-rl/pkg/model"
	g "gorgonia.org/gorgonia"
)

// FC is a fully connected layer of neurons.
type FC struct {
	// Weights for this layer.
	Weights *g.Node

	// activationFn is the activation function for this layer.
	activationFn model.ActivationFn
}

// FCOpts are options for a fully connected layer.
type FCOpts func(*FC)

// NewFC returns a new fully connected layer.
func NewFC(weights *g.Node, opts ...FCOpts) *FC {
	fc := &FC{
		Weights: weights,
	}
	for _, opt := range opts {
		opt(fc)
	}
	return fc
}

// WithActivationFn adds an activation function to the layer.
func WithActivationFn(fn model.ActivationFn) func(*FC) {
	return func(f *FC) {
		f.activationFn = fn
	}
}

// Fwd is a foward pass on a single fully connected layer.
func (f *FC) Fwd(x *g.Node) (*g.Node, error) {
	prod := g.Must(g.Mul(x, f.Weights))
	if f.activationFn == nil {
		return prod, nil
	}
	return f.activationFn(prod)
}

// Learnables are the learnable parameters of the fully connected layer.
func (f *FC) Learnables() g.Nodes {
	return g.Nodes{f.Weights}
}
