package core

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint/internal/config"
	pb "github.com/hashicorp/waypoint/internal/server/gen"
)

// Build builds the artifact from source for this app.
// TODO(mitchellh): test
func (a *App) Build(ctx context.Context, optFuncs ...BuildOption) (
	*pb.Build,
	*pb.PushedArtifact,
	error,
) {
	opts, err := newBuildOptions(optFuncs...)
	if err != nil {
		return nil, nil, err
	}

	// First we do the build
	_, msg, err := a.doOperation(ctx, a.logger.Named("build"), &buildOperation{})
	if err != nil {
		return nil, nil, err
	}
	build := msg.(*pb.Build)

	// If we're not pushing, then we're done!
	if !opts.Push {
		return build, nil, nil
	}

	// We're also pushing to a registry, so invoke that.
	artifact, err := a.PushBuild(ctx, PushWithBuild(build))
	return build, artifact, err
}

// BuildOption is used to configure a Build
type BuildOption func(*buildOptions) error

// BuildWithPush sets whether or not the build will push. The default
// is for the build to push.
func BuildWithPush(v bool) BuildOption {
	return func(opts *buildOptions) error {
		opts.Push = v
		return nil
	}
}

type buildOptions struct {
	Push bool
}

func defaultBuildOptions() *buildOptions {
	return &buildOptions{
		Push: true,
	}
}

func newBuildOptions(opts ...BuildOption) (*buildOptions, error) {
	def := defaultBuildOptions()
	for _, f := range opts {
		if err := f(def); err != nil {
			return nil, err
		}
	}

	return def, def.Validate()
}

func (opts *buildOptions) Validate() error {
	return nil
}

// buildOperation implements the operation interface.
type buildOperation struct {
	Build *pb.Build
}

func (op *buildOperation) Init(app *App) (proto.Message, error) {
	builder, ok := app.components[app.Builder]
	if !ok {
		return nil, status.Error(codes.NotFound, "no builder configured")
	}

	return &pb.Build{
		Application: app.ref,
		Workspace:   app.workspace,
		Component:   builder.Info,
	}, nil
}

func (op *buildOperation) Hooks(app *App) map[string][]*config.Hook {
	builder, ok := app.components[app.Builder]
	if !ok {
		return nil
	}
	return builder.Hooks
}

func (op *buildOperation) Labels(app *App) map[string]string {
	builder, ok := app.components[app.Builder]
	if !ok {
		return nil
	}
	return builder.Labels
}

func (op *buildOperation) Upsert(
	ctx context.Context,
	client pb.WaypointClient,
	msg proto.Message,
) (proto.Message, error) {
	resp, err := client.UpsertBuild(ctx, &pb.UpsertBuildRequest{
		Build: msg.(*pb.Build),
	})
	if err != nil {
		return nil, err
	}

	return resp.Build, nil
}

func (op *buildOperation) Do(ctx context.Context, log hclog.Logger, app *App, _ proto.Message) (interface{}, error) {
	return app.callDynamicFunc(ctx,
		log,
		(*component.Artifact)(nil),
		app.Builder,
		app.Builder.BuildFunc(),
	)
}

func (op *buildOperation) StatusPtr(msg proto.Message) **pb.Status {
	return &(msg.(*pb.Build).Status)
}

func (op *buildOperation) ValuePtr(msg proto.Message) **any.Any {
	v := msg.(*pb.Build)
	if v.Artifact == nil {
		v.Artifact = &pb.Artifact{}
	}

	return &v.Artifact.Artifact
}

var _ operation = (*buildOperation)(nil)
