import { describe, expect, it, vi } from "vitest";
import {
  replaceVideoStreamSlot,
  resetVideoStreams,
  stopMediaStreamTracks
} from "./liveMediaCleanup.js";

function createFakeStream() {
  const tracks = [{ stop: vi.fn() }, { stop: vi.fn() }];
  return {
    getTracks: () => tracks,
    tracks
  };
}

describe("liveMediaCleanup", () => {
  it("stops every track in a stream", () => {
    const stream = createFakeStream();

    stopMediaStreamTracks(stream);

    expect(stream.tracks[0].stop).toHaveBeenCalledTimes(1);
    expect(stream.tracks[1].stop).toHaveBeenCalledTimes(1);
  });

  it("stops the previous slot stream when replacing it", () => {
    const previousRgb = createFakeStream();
    const nextRgb = createFakeStream();
    const audio = createFakeStream();

    const nextStreams = replaceVideoStreamSlot({ rgb: previousRgb, thermal: null, audio }, "rgb", nextRgb);

    expect(nextStreams).toMatchObject({ rgb: nextRgb, thermal: null, audio });
    expect(previousRgb.tracks[0].stop).toHaveBeenCalledTimes(1);
    expect(previousRgb.tracks[1].stop).toHaveBeenCalledTimes(1);
    expect(audio.tracks[0].stop).not.toHaveBeenCalled();
    expect(nextRgb.tracks[0].stop).not.toHaveBeenCalled();
  });

  it("stops all streams when resetting mission streams", () => {
    const rgb = createFakeStream();
    const thermal = createFakeStream();
    const audio = createFakeStream();

    const nextStreams = resetVideoStreams({ rgb, thermal, audio });

    expect(nextStreams).toEqual({ rgb: null, thermal: null, audio: null });
    expect(rgb.tracks[0].stop).toHaveBeenCalledTimes(1);
    expect(thermal.tracks[0].stop).toHaveBeenCalledTimes(1);
    expect(audio.tracks[0].stop).toHaveBeenCalledTimes(1);
  });
});
