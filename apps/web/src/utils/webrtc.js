export function waitForIceGatheringComplete(peerConnection, timeoutMs = 5000) {
  if (peerConnection.iceGatheringState === "complete") {
    return Promise.resolve();
  }

  return new Promise((resolve) => {
    let settled = false;
    const cleanup = () => {
      if (settled) {
        return;
      }
      settled = true;
      window.clearTimeout(timeout);
      peerConnection.removeEventListener("icegatheringstatechange", handleStateChange);
      resolve();
    };
    const handleStateChange = () => {
      if (peerConnection.iceGatheringState === "complete") {
        cleanup();
      }
    };
    const timeout = window.setTimeout(cleanup, timeoutMs);
    peerConnection.addEventListener("icegatheringstatechange", handleStateChange);
  });
}
