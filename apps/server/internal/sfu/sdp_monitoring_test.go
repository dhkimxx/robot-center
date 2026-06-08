package sfu

import (
	"strings"
	"testing"
)

func TestSDPMonitorFieldsSummarizesRobotOfferWithoutRawSDP(t *testing.T) {
	rawSDP := strings.Join([]string{
		"v=0",
		"o=- 1 1 IN IP4 0.0.0.0",
		"s=-",
		"t=0 0",
		"a=group:BUNDLE video0 audio1 application2",
		"m=video 9 UDP/TLS/RTP/SAVPF 96",
		"a=mid:video0",
		"a=rtpmap:96 H264/90000",
		"a=msid:robot-stream track.video_1",
		"m=audio 9 UDP/TLS/RTP/SAVPF 111",
		"a=mid:audio1",
		"a=rtpmap:111 opus/48000/2",
		"a=msid:robot-stream track.audio_1",
		"m=application 0 UDP/DTLS/SCTP webrtc-datachannel",
		"a=mid:application2",
		"a=sctp-port:5000",
		"",
	}, "\r\n")

	fields := monitorFieldsMap(sdpMonitorFields(rawSDP))
	if fields["sdp"] != "parsed" {
		t.Fatalf("expected parsed SDP summary, got %#v", fields)
	}
	if fields["media"] != "video:1,audio:1,application:1" {
		t.Fatalf("unexpected media summary: %#v", fields)
	}
	if fields["bundle"] != true || fields["datachannel"] != true || fields["applicationZeroPort"] != true {
		t.Fatalf("expected bundle/datachannel/zero-port flags, got %#v", fields)
	}
	if fields["codecs"] != "H264,opus" {
		t.Fatalf("unexpected codec summary: %#v", fields)
	}
	if fields["tracks"] != "track.audio_1,track.video_1" {
		t.Fatalf("unexpected track summary: %#v", fields)
	}
}

func monitorFieldsMap(fields []any) map[string]any {
	output := map[string]any{}
	for index := 0; index+1 < len(fields); index += 2 {
		key, ok := fields[index].(string)
		if !ok {
			continue
		}
		output[key] = fields[index+1]
	}
	return output
}
