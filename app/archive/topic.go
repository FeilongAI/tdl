package archive

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
)

type topicManager struct {
	ctx         context.Context
	client      *telegram.Client
	kvd         storage.Storage
	forumPeer   peers.Peer
	chatID      string
	stripPrefix string
	topicCache  map[string]int // folder path -> topic ID
}

func newTopicManager(ctx context.Context, client *telegram.Client, kvd storage.Storage, forumPeer peers.Peer, chatID, stripPrefix string) *topicManager {
	return &topicManager{
		ctx:         ctx,
		client:      client,
		kvd:         kvd,
		forumPeer:   forumPeer,
		chatID:      chatID,
		stripPrefix: stripPrefix,
		topicCache:  make(map[string]int),
	}
}

func (tm *topicManager) ensureTopics(ctx context.Context, folders []string) error {
	// Load existing mappings from KV storage
	if err := tm.loadMappings(); err != nil {
		logctx.From(ctx).Warn("Failed to load topic mappings, will create new ones", zap.Error(err))
	}

	// Process each folder
	for _, folder := range folders {
		if _, exists := tm.topicCache[folder]; exists {
			continue // Already have this topic
		}

		// Create new topic
		relPath, err := getRelativePath(folder, tm.stripPrefix)
		if err != nil {
			return errors.Wrap(err, "get relative path")
		}

		topicID, err := tm.createTopic(ctx, relPath)
		if err != nil {
			return errors.Wrapf(err, "create topic for folder: %s", relPath)
		}

		tm.topicCache[folder] = topicID

		// Save mapping
		if err := tm.saveMapping(folder, topicID); err != nil {
			logctx.From(ctx).Warn("Failed to save topic mapping", zap.Error(err))
		}

		color.Green("Created topic: %s (ID: %d)", relPath, topicID)
	}

	return nil
}

func (tm *topicManager) createTopic(ctx context.Context, title string) (int, error) {
	// Get input channel
	inputChannel, ok := tm.forumPeer.InputPeer().(*tg.InputPeerChannel)
	if !ok {
		return 0, errors.New("forum peer is not a channel")
	}

	// Create forum topic request
	req := &tg.ChannelsCreateForumTopicRequest{
		Channel:  &tg.InputChannel{ChannelID: inputChannel.ChannelID, AccessHash: inputChannel.AccessHash},
		Title:    title,
		RandomID: rand.Int63(),
	}

	// Execute request
	updates, err := tm.client.API().ChannelsCreateForumTopic(ctx, req)
	if err != nil {
		return 0, errors.Wrap(err, "create forum topic")
	}

	// Extract topic ID from updates
	topicID, err := extractTopicID(updates)
	if err != nil {
		return 0, errors.Wrap(err, "extract topic ID")
	}

	return topicID, nil
}

func (tm *topicManager) getTopicID(folder string) (int, error) {
	topicID, exists := tm.topicCache[folder]
	if !exists {
		return 0, fmt.Errorf("topic not found for folder: %s", folder)
	}
	return topicID, nil
}

func (tm *topicManager) loadMappings() error {
	data, err := tm.kvd.Get(tm.ctx, tm.getMappingKey())
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &tm.topicCache)
}

func (tm *topicManager) saveMapping(folder string, topicID int) error {
	data, err := json.Marshal(tm.topicCache)
	if err != nil {
		return err
	}

	return tm.kvd.Set(tm.ctx, tm.getMappingKey(), data)
}

func (tm *topicManager) getMappingKey() string {
	return fmt.Sprintf("archive:topics:%s", tm.chatID)
}

func extractTopicID(updates tg.UpdatesClass) (int, error) {
	switch u := updates.(type) {
	case *tg.Updates:
		for _, update := range u.Updates {
			if msg, ok := update.(*tg.UpdateNewChannelMessage); ok {
				if service, ok := msg.Message.(*tg.MessageService); ok {
					if _, ok := service.Action.(*tg.MessageActionTopicCreate); ok {
						return service.ID, nil
					}
				}
			}
		}
	}

	return 0, errors.New("topic ID not found in updates")
}
