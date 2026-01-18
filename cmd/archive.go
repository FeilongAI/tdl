package cmd

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/archive"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
)

func NewArchive() *cobra.Command {
	var opts archive.Options

	cmd := &cobra.Command{
		Use:     "archive",
		Short:   "Archive files to Telegram forum with folder tree structure",
		Long:    `Archive local files to a Telegram forum group, creating topics for each folder in the directory tree`,
		GroupID: groupTools.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tRun(cmd.Context(), func(ctx context.Context, c *telegram.Client, kvd storage.Storage) error {
				return archive.Run(logctx.Named(ctx, "archive"), c, kvd, opts)
			})
		},
	}

	const (
		_chat        = "chat"
		path         = "path"
		stripPrefix  = "strip-prefix"
		include      = "include"
		exclude      = "exclude"
	)

	cmd.Flags().StringVarP(&opts.Chat, _chat, "c", "", "forum group id or domain (required)")
	cmd.Flags().StringVarP(&opts.Path, path, "p", "", "local folder path to archive (required)")
	cmd.Flags().StringVar(&opts.StripPrefix, stripPrefix, "", "path prefix to strip (default: parent of path)")
	cmd.Flags().StringSliceVarP(&opts.Includes, include, "i", []string{}, "include the specified file extensions")
	cmd.Flags().StringSliceVarP(&opts.Excludes, exclude, "e", []string{}, "exclude the specified file extensions")
	cmd.Flags().BoolVar(&opts.Remove, "rm", false, "remove the uploaded files after uploading")
	cmd.Flags().BoolVar(&opts.Photo, "photo", false, "upload images as photos instead of files")

	// mark required flags
	_ = cmd.MarkFlagRequired(_chat)
	_ = cmd.MarkFlagRequired(path)

	// mark mutually exclusive flags
	cmd.MarkFlagsMutuallyExclusive(include, exclude)

	return cmd
}