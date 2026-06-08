package sfu

import (
	"sort"
	"strconv"
	"strings"

	"github.com/pion/sdp/v3"
)

func sdpMonitorFields(rawSDP string) []any {
	var sessionDescription sdp.SessionDescription
	if err := sessionDescription.UnmarshalString(rawSDP); err != nil {
		return []any{"sdp", "parse_failed"}
	}

	mediaCounts := map[string]int{}
	codecs := map[string]struct{}{}
	trackIDs := map[string]struct{}{}
	hasDataChannel := false
	hasZeroPortApplication := false

	for _, mediaDescription := range sessionDescription.MediaDescriptions {
		if mediaDescription == nil {
			continue
		}
		mediaKind := strings.TrimSpace(mediaDescription.MediaName.Media)
		if mediaKind == "" {
			mediaKind = "unknown"
		}
		mediaCounts[mediaKind]++
		if mediaKind == "application" && mediaDescription.MediaName.Port.Value == sdpInactivePort {
			hasZeroPortApplication = true
		}
		if isDataChannelMedia(mediaDescription) {
			hasDataChannel = true
		}
		for _, attribute := range mediaDescription.Attributes {
			switch attribute.Key {
			case "msid":
				fields := strings.Fields(attribute.Value)
				if len(fields) >= 2 {
					trackIDs[fields[1]] = struct{}{}
				}
			case "rtpmap":
				if codec := codecNameFromRTPMap(attribute.Value); codec != "" {
					codecs[codec] = struct{}{}
				}
			}
		}
	}

	return []any{
		"sdp", "parsed",
		"media", formatSDPMediaCounts(mediaCounts),
		"bundle", len(bundledMediaIdentifiers(sessionDescription)) > 0,
		"datachannel", hasDataChannel,
		"applicationZeroPort", hasZeroPortApplication,
		"codecs", joinSortedKeys(codecs, 8),
		"tracks", joinSortedKeys(trackIDs, 8),
	}
}

func isDataChannelMedia(mediaDescription *sdp.MediaDescription) bool {
	if mediaDescription == nil || mediaDescription.MediaName.Media != "application" {
		return false
	}
	if !containsString(mediaDescription.MediaName.Formats, "webrtc-datachannel") {
		return false
	}
	if !containsString(mediaDescription.MediaName.Protos, "SCTP") {
		return false
	}
	_, hasSCTPPort := mediaDescription.Attribute("sctp-port")
	return hasSCTPPort
}

func codecNameFromRTPMap(value string) string {
	fields := strings.Fields(value)
	if len(fields) < 2 {
		return ""
	}
	codecFields := strings.Split(fields[1], "/")
	if len(codecFields) == 0 {
		return ""
	}
	return strings.TrimSpace(codecFields[0])
}

func formatSDPMediaCounts(mediaCounts map[string]int) string {
	order := []string{"video", "audio", "application"}
	parts := make([]string, 0, len(mediaCounts))
	seen := map[string]struct{}{}
	for _, mediaKind := range order {
		count := mediaCounts[mediaKind]
		if count == 0 {
			continue
		}
		parts = append(parts, mediaKind+":"+strconv.Itoa(count))
		seen[mediaKind] = struct{}{}
	}
	remaining := make([]string, 0)
	for mediaKind := range mediaCounts {
		if _, ok := seen[mediaKind]; ok {
			continue
		}
		remaining = append(remaining, mediaKind)
	}
	sort.Strings(remaining)
	for _, mediaKind := range remaining {
		parts = append(parts, mediaKind+":"+strconv.Itoa(mediaCounts[mediaKind]))
	}
	return strings.Join(parts, ",")
}

func joinSortedKeys(values map[string]struct{}, limit int) string {
	if len(values) == 0 {
		return ""
	}
	output := make([]string, 0, len(values))
	for value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			output = append(output, value)
		}
	}
	sort.Strings(output)
	if limit > 0 && len(output) > limit {
		output = output[:limit]
		output = append(output, "more")
	}
	return strings.Join(output, ",")
}
