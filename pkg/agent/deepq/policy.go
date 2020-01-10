package deepq

import (
	"github.com/pbarker/go-rl/pkg/model"
	"github.com/pbarker/go-rl/pkg/model/layers"
	g "gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

// // Policy for a dqn network.
// type Policy interface {
// 	// Step takes a step given the current observation.
// 	Step(observation *tensor.Dense) (actions, qValues, states *tensor.Dense, err error)
// }

// Policy is a dqn policy using a fully connected feed forward neural network.
type Policy struct {
	graph    *g.ExprGraph
	inputs   *g.Node
	expected *g.Node
	chain    *model.Chain

	pred    *g.Node
	predVal g.Value
}

// PolicyConfig is the configuration for a FCPolicy.
type PolicyConfig struct {
	// BatchSize is the size of the batch used to train.
	BatchSize int

	// Type of the network.
	Type tensor.Dtype

	// Cost function to evaluate network perfomance.
	CostFn model.CostFn

	// ChainBuilder is a builder for a chain for layers.
	ChainBuilder ChainBuilder
}

// DefaultPolicyConfig is the default configuration for and FCPolicy.
var DefaultPolicyConfig = &PolicyConfig{
	BatchSize:    100,
	Type:         tensor.Float32,
	CostFn:       model.MeanSquaredError,
	ChainBuilder: DefaultFCChainBuilder,
}

// ChainBuilder is a builder of layer chains for RL puroposes.
type ChainBuilder func(graph *g.ExprGraph, actionSpaceSize int) *model.Chain

// DefaultFCChainBuilder creates a default fully connected network for the given action space size.
func DefaultFCChainBuilder(graph *g.ExprGraph, actionSpaceSize int) *model.Chain {
	chain := model.NewChain(
		layers.NewFC(g.NewMatrix(graph, g.Float32, g.WithShape(actionSpaceSize, 2), g.WithName("w0"), g.WithInit(g.GlorotU(1.0))), layers.WithActivationFn(g.Tanh)),
		layers.NewFC(g.NewMatrix(graph, g.Float32, g.WithShape(2, 100), g.WithName("w1"), g.WithInit(g.GlorotU(1.0))), layers.WithActivationFn(g.Tanh)),
		layers.NewFC(g.NewMatrix(graph, g.Float32, g.WithShape(100, 100), g.WithName("w2"), g.WithInit(g.GlorotU(1.0))), layers.WithActivationFn(g.Tanh)),
		layers.NewFC(g.NewMatrix(graph, g.Float32, g.WithShape(100, 1), g.WithName("w2"), g.WithInit(g.GlorotU(1.0)))),
	)
	return chain
}

// NewPolicy creates a new feed forward policy.
func NewPolicy(c *PolicyConfig, actionSpaceSize int) (*Policy, error) {
	graph := g.NewGraph()

	inputs := g.NewMatrix(graph, g.Float32, g.WithShape(c.BatchSize, actionSpaceSize), g.WithName("inputs"), g.WithInit(g.Zeroes()))
	expected := g.NewVector(graph, g.Float32, g.WithShape(c.BatchSize), g.WithName("expected"), g.WithInit(g.Zeroes()))

	chain := c.ChainBuilder(graph, actionSpaceSize)
	prediction, err := chain.Fwd(inputs)
	if err != nil {
		return nil, err
	}
	cost := c.CostFn(prediction, expected)
	if _, err = g.Grad(cost, chain.Learnables()...); err != nil {
		return nil, err
	}

	return &Policy{
		graph:    graph,
		inputs:   inputs,
		expected: expected,
		chain:    chain,
	}, nil
}

// Step takes a step given the current observation.
func (p *Policy) Step(observation *tensor.Dense) (actions, qValues, states *tensor.Dense, err error) {

}
