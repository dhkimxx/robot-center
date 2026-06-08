package sfu

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

type localCandidateBuffer struct {
	mu         sync.Mutex
	candidates []webrtc.ICECandidateInit
}

func (b *localCandidateBuffer) append(candidate *webrtc.ICECandidate) {
	if candidate == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.candidates = append(b.candidates, candidate.ToJSON())
}

func (b *localCandidateBuffer) drain() []webrtc.ICECandidateInit {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.candidates) == 0 {
		return nil
	}
	candidates := append([]webrtc.ICECandidateInit(nil), b.candidates...)
	b.candidates = nil
	return candidates
}
