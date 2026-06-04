package sfu

import (
	"strings"
	"time"
)

func (h *Hub) emitPublisherEvent(event PublisherEvent) {
	if h.config.OnPublisherEvent == nil {
		return
	}
	event.RoomID = strings.TrimSpace(event.RoomID)
	event.RobotCode = strings.TrimSpace(event.RobotCode)
	event.PublisherPeerID = strings.TrimSpace(event.PublisherPeerID)
	event.Reason = strings.TrimSpace(event.Reason)
	if event.ObservedAt.IsZero() {
		event.ObservedAt = time.Now().UTC()
	} else {
		event.ObservedAt = event.ObservedAt.UTC()
	}
	go h.config.OnPublisherEvent(event)
}

func publisherEndedEvent(roomID string, publisher *publisherSession, reason string, observedAt time.Time) (PublisherEvent, bool) {
	if publisher == nil || strings.TrimSpace(publisher.peerID) == "" {
		return PublisherEvent{}, false
	}
	return PublisherEvent{
		Type:            PublisherEventEnded,
		RoomID:          roomID,
		RobotCode:       publisher.robotCode,
		PublisherPeerID: publisher.peerID,
		Reason:          reason,
		ObservedAt:      observedAt,
	}, true
}
