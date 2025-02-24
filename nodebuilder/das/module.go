package das

import (
	"context"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/das"
	"github.com/celestiaorg/celestia-node/fraud"
	fraudServ "github.com/celestiaorg/celestia-node/nodebuilder/fraud"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	err := cfg.Validate()

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(err),
		fx.Provide(
			func(c Config) []das.Option {
				return []das.Option{
					das.WithSamplingRange(c.SamplingRange),
					das.WithConcurrencyLimit(c.ConcurrencyLimit),
					das.WithPriorityQueueSize(c.PriorityQueueSize),
					das.WithBackgroundStoreInterval(c.BackgroundStoreInterval),
					das.WithSampleFrom(c.SampleFrom),
				}
			},
		),
	)

	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"daser",
			baseComponents,
			fx.Provide(fx.Annotate(
				NewDASer,
				fx.OnStart(func(startCtx, ctx context.Context, fservice fraud.Service, das *das.DASer) error {
					return fraudServ.Lifecycle(startCtx, ctx, fraud.BadEncoding, fservice,
						das.Start, das.Stop)
				}),
				fx.OnStop(func(ctx context.Context, das *das.DASer) error {
					return das.Stop(ctx)
				}),
			)),
			// Module is needed for the RPC handler
			fx.Provide(func(das *das.DASer) Module {
				return das
			}),
		)
	case node.Bridge:
		return fx.Module(
			"daser",
			baseComponents,
			fx.Provide(newDaserStub),
		)
	default:
		panic("invalid node type")
	}
}
