import { makeMissionRobotKey } from "../missions/missionHelpers.js";
import { LiveSessionStatus } from "./liveConnectionStates.js";
import { createEmptyDetectionOverlays } from "./liveEventStrategies.js";
import { createEmptyVideoStreams } from "./liveMediaCleanup.js";

export function createEmptyLiveSession() {
  return {
    attemptId: null,
    events: [],
    detectionOverlays: createEmptyDetectionOverlays(),
    sensor: null,
    status: LiveSessionStatus.DISCONNECTED,
    telemetry: null,
    videoStreams: createEmptyVideoStreams()
  };
}

export function makeLiveTargetKey(target) {
  return target ? makeMissionRobotKey(target.mission.missionCode, target.robotCode) : "";
}

export function findRobotCodeForRemoteTrack(event, missionTargets) {
  const raw = [
    event.track?.id,
    event.track?.label,
    ...(event.streams ?? []).flatMap((stream) => [stream.id, stream.label])
  ].filter(Boolean).join(" ").toLowerCase();
  return [...missionTargets]
    .sort((left, right) => right.robotCode.length - left.robotCode.length)
    .find((target) => raw.includes(target.robotCode.toLowerCase()))?.robotCode ?? "";
}

export function findTrackSlot(event) {
  const raw = [
    event.track?.id,
    event.track?.label,
    ...(event.streams ?? []).flatMap((stream) => [stream.id, stream.label])
  ].filter(Boolean).join(" ").toLowerCase();
  if (raw.includes("track.video_2")) {
    return "thermal";
  }
  if (raw.includes("track.video_1")) {
    return "rgb";
  }
  if (raw.includes("track.audio_1") || raw.includes("track.audio_2")) {
    return "audio";
  }
  return "unmapped";
}

export function describeRemoteTrack(event) {
  const trackId = event.track?.id || "-";
  const streamIds = (event.streams ?? []).map((stream) => stream.id).filter(Boolean).join(",");
  const kind = event.track?.kind || "-";
  return `kind=${kind} trackId=${trackId} streamIds=${streamIds || "-"}`;
}

export function findRobotCodeFromDataMessage(message) {
  try {
    const parsed = JSON.parse(message);
    return parsed?.robotCode
      ?? parsed?.payload?.robotCode
      ?? "";
  } catch {
    return "";
  }
}

export function formatMediaChannelCount(streamSource) {
  const mediaCount = streamSource?.trackCount
    ?? streamSource?.tracks?.length
    ?? streamSource?.publishedTracks?.length
    ?? 0;
  const dataChannelCount = streamSource?.dataChannelCount
    ?? streamSource?.dataChannels?.length
    ?? streamSource?.publishedDataChannels?.length
    ?? 0;
  if (mediaCount > 0 && dataChannelCount > 0) {
    return `영상/음성 ${mediaCount}채널 / 센서 ${dataChannelCount}채널`;
  }
  if (mediaCount > 0) {
    return `영상/음성 ${mediaCount}채널`;
  }
  if (dataChannelCount > 0) {
    return `센서 ${dataChannelCount}채널`;
  }
  return "송출 대기";
}

export function getStreamingSubscriberCount(streamSource) {
  return streamSource?.subscriberCount
    ?? streamSource?.operatorCount
    ?? streamSource?.browserCount
    ?? streamSource?.viewerCount
    ?? null;
}

export function formatStreamingSubscriberCount(streamSource) {
  const subscriberCount = getStreamingSubscriberCount(streamSource);
  return subscriberCount === null ? "관제 -" : `관제 ${subscriberCount}명`;
}
