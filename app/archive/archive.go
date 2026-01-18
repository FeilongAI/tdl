package archive

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/spf13/viper"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
)

type Options struct {
	Chat        string
	Path        string
	StripPrefix string
	Includes    []string
	Excludes    []string
	Remove      bool
	Photo       bool
}

func Run(ctx context.Context, c *telegram.Client, kvd storage.Storage, opts Options) (rerr error) {
	// Normalize paths
	opts.Path = filepath.Clean(opts.Path)
	if opts.StripPrefix == "" {
		opts.StripPrefix = filepath.Dir(opts.Path)
	}
	opts.StripPrefix = filepath.Clean(opts.StripPrefix)

	color.Blue("Archive path: %s", opts.Path)
	color.Blue("Strip prefix: %s", opts.StripPrefix)

	// Create DC pool
	pool := dcpool.NewPool(c,
		int64(viper.GetInt(consts.FlagPoolSize)),
		tclient.NewDefaultMiddlewares(ctx, viper.GetDuration(consts.FlagReconnectTimeout))...)
	defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

	// Initialize peers manager
	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Default(ctx))

	// Get forum channel
	forumPeer, err := tutil.GetInputPeer(ctx, manager, opts.Chat)
	if err != nil {
		return errors.Wrap(err, "get forum chat")
	}

	// Scan folder tree
	tree, err := scanFolderTree(opts.Path, opts.Includes, opts.Excludes)
	if err != nil {
		return errors.Wrap(err, "scan folder tree")
	}

	color.Blue("Total folders: %d, Total files: %d", len(tree.Folders), tree.TotalFiles)

	// Create or get topics for each folder
	topicManager := newTopicManager(ctx, c, kvd, forumPeer, opts.Chat, opts.StripPrefix)
	if err := topicManager.ensureTopics(ctx, tree.Folders); err != nil {
		return errors.Wrap(err, "ensure topics")
	}

	color.Green("Topics created/loaded successfully!")

	// Prepare upload items
	items, err := prepareUploadItems(tree, topicManager, opts.Photo)
	if err != nil {
		return errors.Wrap(err, "prepare upload items")
	}

	if len(items) == 0 {
		color.Yellow("No files to upload")
		return nil
	}

	color.Blue("Starting upload of %d files...", len(items))

	// Setup progress tracker
	upProgress := prog.New(utils.Byte.FormatBinaryBytes)
	upProgress.SetNumTrackersExpected(len(items))
	prog.EnablePS(ctx, upProgress)

	// Create uploader
	uploaderOpts := uploader.Options{
		Client:   pool.Default(ctx),
		Threads:  viper.GetInt(consts.FlagThreads),
		Iter:     newIter(items, forumPeer, opts.Remove, viper.GetDuration(consts.FlagDelay)),
		Progress: newProgress(upProgress),
	}

	up := uploader.New(uploaderOpts)

	go upProgress.Render()
	defer prog.Wait(ctx, upProgress)

	if err := up.Upload(ctx, viper.GetInt(consts.FlagLimit)); err != nil {
		return errors.Wrap(err, "upload files")
	}

	color.Green("Archive completed successfully!")
	return nil
}

func getRelativePath(fullPath, stripPrefix string) (string, error) {
	rel, err := filepath.Rel(stripPrefix, fullPath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %s is not under %s", fullPath, stripPrefix)
	}
	return rel, nil
}