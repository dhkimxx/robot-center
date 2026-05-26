export function createEmptyVideoStreams() {
  return { rgb: null, thermal: null, audio: null };
}

export function stopMediaTrack(track) {
  if (!track || typeof track.stop !== "function") {
    return;
  }
  try {
    track.stop();
  } catch {
    // Browser cleanup must be best-effort; a stopped track should not break route changes.
  }
}

export function stopMediaStreamTracks(stream) {
  if (!stream || typeof stream.getTracks !== "function") {
    return;
  }
  stream.getTracks().forEach(stopMediaTrack);
}

export function stopVideoStreams(videoStreams = {}) {
  Object.values(videoStreams).forEach(stopMediaStreamTracks);
}

export function replaceVideoStreamSlot(videoStreams = createEmptyVideoStreams(), slot, stream) {
  const previousStream = videoStreams[slot] ?? null;
  if (previousStream && previousStream !== stream) {
    stopMediaStreamTracks(previousStream);
  }
  return { ...videoStreams, [slot]: stream };
}

export function resetVideoStreams(videoStreams = createEmptyVideoStreams()) {
  stopVideoStreams(videoStreams);
  return createEmptyVideoStreams();
}
