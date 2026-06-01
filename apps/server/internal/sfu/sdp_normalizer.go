package sfu

import (
	"strings"

	"github.com/pion/sdp/v3"
)

const sdpInactivePort = 0
const sdpIcePlaceholderPort = 9

type sdpLine struct {
	content string
	ending  string
}

func normalizeBundledApplicationDataChannelSDP(rawSDP string) (string, bool) {
	var sessionDescription sdp.SessionDescription
	if err := sessionDescription.UnmarshalString(rawSDP); err != nil {
		return rawSDP, false
	}

	bundleMIDs := bundledMediaIdentifiers(sessionDescription)
	if len(bundleMIDs) == 0 {
		return rawSDP, false
	}

	targetMIDs := map[string]struct{}{}
	for _, mediaDescription := range sessionDescription.MediaDescriptions {
		if mediaDescription == nil || !isZeroPortBundledDataChannelMedia(mediaDescription, bundleMIDs) {
			continue
		}
		mid, _ := mediaDescription.Attribute("mid")
		targetMIDs[strings.TrimSpace(mid)] = struct{}{}
	}
	if len(targetMIDs) == 0 {
		return rawSDP, false
	}

	return rewriteApplicationMediaPorts(rawSDP, targetMIDs)
}

func bundledMediaIdentifiers(sessionDescription sdp.SessionDescription) map[string]struct{} {
	mids := map[string]struct{}{}
	for _, attribute := range sessionDescription.Attributes {
		if attribute.Key != "group" {
			continue
		}
		fields := strings.Fields(attribute.Value)
		if len(fields) < 2 || fields[0] != "BUNDLE" {
			continue
		}
		for _, mid := range fields[1:] {
			mids[mid] = struct{}{}
		}
	}
	return mids
}

func isZeroPortBundledDataChannelMedia(mediaDescription *sdp.MediaDescription, bundleMIDs map[string]struct{}) bool {
	if mediaDescription.MediaName.Media != "application" {
		return false
	}
	if mediaDescription.MediaName.Port.Value != sdpInactivePort || mediaDescription.MediaName.Port.Range != nil {
		return false
	}
	if !containsString(mediaDescription.MediaName.Formats, "webrtc-datachannel") {
		return false
	}
	if !containsString(mediaDescription.MediaName.Protos, "SCTP") {
		return false
	}
	if _, ok := mediaDescription.Attribute("sctp-port"); !ok {
		return false
	}
	mid, ok := mediaDescription.Attribute("mid")
	if !ok {
		return false
	}
	_, bundled := bundleMIDs[strings.TrimSpace(mid)]
	return bundled
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), expected) {
			return true
		}
	}
	return false
}

func rewriteApplicationMediaPorts(rawSDP string, targetMIDs map[string]struct{}) (string, bool) {
	lines := splitSDPLines(rawSDP)
	sectionStart := -1
	sectionMID := ""
	changed := false

	finalizeSection := func() {
		if sectionStart < 0 {
			return
		}
		if _, ok := targetMIDs[sectionMID]; !ok {
			return
		}
		updated, ok := rewriteApplicationMediaPort(lines[sectionStart].content)
		if !ok {
			return
		}
		lines[sectionStart].content = updated
		changed = true
	}

	for index := range lines {
		content := lines[index].content
		switch {
		case strings.HasPrefix(content, "m="):
			finalizeSection()
			sectionStart = index
			sectionMID = ""
		case sectionStart >= 0 && strings.HasPrefix(content, "a=mid:"):
			sectionMID = strings.TrimSpace(strings.TrimPrefix(content, "a=mid:"))
		}
	}
	finalizeSection()

	if !changed {
		return rawSDP, false
	}
	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString(line.content)
		builder.WriteString(line.ending)
	}
	return builder.String(), true
}

func splitSDPLines(rawSDP string) []sdpLine {
	parts := strings.SplitAfter(rawSDP, "\n")
	lines := make([]sdpLine, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		line := sdpLine{content: part}
		if strings.HasSuffix(line.content, "\n") {
			line.ending = "\n"
			line.content = strings.TrimSuffix(line.content, "\n")
		}
		if strings.HasSuffix(line.content, "\r") {
			line.ending = "\r" + line.ending
			line.content = strings.TrimSuffix(line.content, "\r")
		}
		lines = append(lines, line)
	}
	return lines
}

func rewriteApplicationMediaPort(line string) (string, bool) {
	if !strings.HasPrefix(line, "m=") {
		return line, false
	}
	fields := strings.Fields(strings.TrimPrefix(line, "m="))
	if len(fields) < 4 || fields[0] != "application" || fields[1] != "0" {
		return line, false
	}
	fields[1] = "9"
	return "m=" + strings.Join(fields, " "), true
}
