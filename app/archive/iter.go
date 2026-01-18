package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/core/util/mediautil"
)

type uploaderFile struct {
	*os.File
	size int64
}

func (f *uploaderFile) Name() string {
	return filepath.Base(f.File.Name())
}

func (f *uploaderFile) Size() int64 {
	return f.size
}

type iterElem struct {
	file    *uploaderFile
	thumb   *uploaderFile
	to      peers.Peer
	caption *entity.Builder
	thread  int

	asPhoto bool
	remove  bool
}

func (e *iterElem) File() uploader.File {
	return e.file
}

func (e *iterElem) Thumb() (uploader.File, bool) {
	if e.thumb == nil {
		return nil, false
	}
	return e.thumb, true
}

func (e *iterElem) Caption() (string, []tg.MessageEntityClass) {
	return e.caption.Complete()
}

func (e *iterElem) To() tg.InputPeerClass {
	return e.to.InputPeer()
}

func (e *iterElem) Thread() int {
	return e.thread
}

func (e *iterElem) AsPhoto() bool {
	return e.asPhoto
}

type iter struct {
	items     []UploadItem
	forumPeer peers.Peer
	remove    bool
	delay     time.Duration

	cur  int
	err  error
	elem uploader.Elem
}

func newIter(items []UploadItem, forumPeer peers.Peer, remove bool, delay time.Duration) *iter {
	return &iter{
		items:     items,
		forumPeer: forumPeer,
		remove:    remove,
		delay:     delay,
		cur:       0,
		err:       nil,
		elem:      nil,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	if i.cur >= len(i.items) || i.err != nil {
		return false
	}

	// Delay between uploads
	if i.delay > 0 && i.cur > 0 {
		time.Sleep(i.delay)
	}

	cur := i.items[i.cur]
	i.cur++

	elem, err := i.processItem(ctx, cur)
	if err != nil {
		i.err = err
		return false
	}

	i.elem = elem
	return true
}

func (i *iter) processItem(ctx context.Context, item UploadItem) (*iterElem, error) {
	// Open file
	file, err := os.Open(item.File.Path)
	if err != nil {
		return nil, errors.Wrap(err, "open file")
	}

	fileWrapper := &uploaderFile{
		File: file,
		size: item.File.Size,
	}

	// Open thumbnail if exists
	var thumb *uploaderFile
	if item.File.Thumb != "" {
		thumbFile, err := os.Open(item.File.Thumb)
		if err != nil {
			return nil, errors.Wrap(err, "open thumbnail")
		}

		// Validate thumbnail
		mime, err := mimetype.DetectFile(item.File.Thumb)
		if err == nil && mediautil.IsImage(mime.String()) {
			thumb = &uploaderFile{
				File: thumbFile,
				size: 0,
			}
		} else {
			thumbFile.Close()
		}
	}

	// Generate caption
	caption := &entity.Builder{}
	fileName := filepath.Base(item.File.Path)
	ext := filepath.Ext(item.File.Path)
	name := strings.TrimSuffix(fileName, ext)

	mime, _ := mimetype.DetectFile(item.File.Path)
	mimeType := ""
	if mime != nil {
		mimeType = mime.String()
	}

	captionText := fmt.Sprintf("<code>%s</code> - <code>%s</code>", name, mimeType)
	if err := html.HTML(strings.NewReader(captionText), caption, html.Options{}); err != nil {
		captionText = name // Fallback to plain text
		caption = &entity.Builder{}
		caption.Plain(captionText)
	}

	return &iterElem{
		file:    fileWrapper,
		thumb:   thumb,
		to:      i.forumPeer,
		caption: caption,
		thread:  item.TopicID,
		asPhoto: item.AsPhoto,
		remove:  i.remove,
	}, nil
}

func (i *iter) Value() uploader.Elem {
	return i.elem
}

func (i *iter) Err() error {
	return i.err
}