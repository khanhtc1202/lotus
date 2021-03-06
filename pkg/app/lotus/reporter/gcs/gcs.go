package gcs

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/model"
	"github.com/nghialv/lotus/pkg/app/lotus/reporter"
)

type builder struct {
}

func NewBuilder() reporter.Builder {
	return &builder{}
}

func (b *builder) Build(r *config.Receiver, opts reporter.BuildOptions) (reporter.Reporter, error) {
	configs, ok := r.Type.(*config.Receiver_Gcs)
	if !ok {
		return nil, fmt.Errorf("wrong receiver type for gcs: %T", r.Type)
	}
	return &gcs{
		bucket:          configs.Gcs.Bucket,
		credentialsFile: r.CredentialsFile(configs.Gcs.Credentials.File),
		logger:          opts.NamedLogger("gcs-reporter"),
	}, nil
}

type gcs struct {
	bucket          string
	credentialsFile string
	logger          *zap.Logger
}

func (g *gcs) Report(ctx context.Context, result *model.Result) (lastErr error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(g.credentialsFile))
	if err != nil {
		g.logger.Error("failed to create gcs storage client", zap.Error(err))
		lastErr = err
		return
	}
	cases := []struct {
		format      model.RenderFormat
		extension   string
		contentType string
	}{
		{
			format:      model.RenderFormatText,
			extension:   "txt",
			contentType: "text/plain",
		},
		{
			format:      model.RenderFormatJson,
			extension:   "json",
			contentType: "application/json",
		},
	}
	for _, c := range cases {
		data, err := result.Render(c.format)
		if err != nil {
			g.logger.Error("failed to render result", zap.Error(err))
			lastErr = err
			continue
		}
		filename := fmt.Sprintf("%s/%s.%s", result.TestID, result.TestID, c.extension)
		g.logger.Info("writing test result to gcs storage",
			zap.String("testID", result.TestID),
			zap.String("filename", filename),
			zap.String("bucket", g.bucket),
		)
		wc := client.Bucket(g.bucket).Object(filename).NewWriter(ctx)
		wc.ContentType = c.contentType
		wc.ACL = []storage.ACLRule{{
			Entity: storage.AllUsers,
			Role:   storage.RoleReader,
		}}
		if _, err := wc.Write(data); err != nil {
			g.logger.Error("failed to write result", zap.Error(err))
			lastErr = err
			continue
		}
		if err := wc.Close(); err != nil {
			g.logger.Error("failed to close writer", zap.Error(err))
			lastErr = err
			continue
		}
	}
	return
}
