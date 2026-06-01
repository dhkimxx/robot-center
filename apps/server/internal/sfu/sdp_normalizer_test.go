package sfu

import (
	"strings"
	"testing"
)

func TestNormalizeBundledApplicationDataChannelSDPRewritesZeroPortApplication(t *testing.T) {
	rawSDP := strings.Join([]string{
		"v=0",
		"o=- 1 1 IN IP4 0.0.0.0",
		"s=-",
		"t=0 0",
		"a=group:BUNDLE audio0 video1 application2",
		"m=audio 9 UDP/TLS/RTP/SAVPF 111",
		"a=mid:audio0",
		"m=video 0 UDP/TLS/RTP/SAVPF 96",
		"a=mid:video1",
		"m=application 0 UDP/DTLS/SCTP webrtc-datachannel",
		"a=mid:application2",
		"a=sctp-port:5000",
		"",
	}, "\r\n")

	normalized, changed := normalizeBundledApplicationDataChannelSDP(rawSDP)
	if !changed {
		t.Fatal("expected SDP to be normalized")
	}
	if !strings.Contains(normalized, "m=application 9 UDP/DTLS/SCTP webrtc-datachannel") {
		t.Fatalf("expected application m-line port to be rewritten, got:\n%s", normalized)
	}
	if !strings.Contains(normalized, "m=video 0 UDP/TLS/RTP/SAVPF 96") {
		t.Fatalf("video m-line must not be rewritten, got:\n%s", normalized)
	}
}

func TestNormalizeBundledApplicationDataChannelSDPKeepsAlreadyActiveApplication(t *testing.T) {
	rawSDP := strings.Join([]string{
		"v=0",
		"o=- 1 1 IN IP4 0.0.0.0",
		"s=-",
		"t=0 0",
		"a=group:BUNDLE audio0 application1",
		"m=audio 9 UDP/TLS/RTP/SAVPF 111",
		"a=mid:audio0",
		"m=application 9 UDP/DTLS/SCTP webrtc-datachannel",
		"a=mid:application1",
		"a=sctp-port:5000",
		"",
	}, "\r\n")

	normalized, changed := normalizeBundledApplicationDataChannelSDP(rawSDP)
	if changed {
		t.Fatalf("expected SDP to remain unchanged, got:\n%s", normalized)
	}
}

func TestNormalizeBundledApplicationDataChannelSDPKeepsNonBundledApplication(t *testing.T) {
	rawSDP := strings.Join([]string{
		"v=0",
		"o=- 1 1 IN IP4 0.0.0.0",
		"s=-",
		"t=0 0",
		"a=group:BUNDLE audio0",
		"m=audio 9 UDP/TLS/RTP/SAVPF 111",
		"a=mid:audio0",
		"m=application 0 UDP/DTLS/SCTP webrtc-datachannel",
		"a=mid:application1",
		"a=sctp-port:5000",
		"",
	}, "\r\n")

	normalized, changed := normalizeBundledApplicationDataChannelSDP(rawSDP)
	if changed {
		t.Fatalf("expected non-bundled application m-line to remain unchanged, got:\n%s", normalized)
	}
}

func TestNormalizeBundledApplicationDataChannelSDPKeepsApplicationWithoutSCTPPort(t *testing.T) {
	rawSDP := strings.Join([]string{
		"v=0",
		"o=- 1 1 IN IP4 0.0.0.0",
		"s=-",
		"t=0 0",
		"a=group:BUNDLE audio0 application1",
		"m=audio 9 UDP/TLS/RTP/SAVPF 111",
		"a=mid:audio0",
		"m=application 0 UDP/DTLS/SCTP webrtc-datachannel",
		"a=mid:application1",
		"",
	}, "\r\n")

	normalized, changed := normalizeBundledApplicationDataChannelSDP(rawSDP)
	if changed {
		t.Fatalf("expected application m-line without sctp-port to remain unchanged, got:\n%s", normalized)
	}
}
