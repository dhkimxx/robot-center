import { makeMissionRobotKey } from "../missions/missionHelpers.js";
import { LiveSessionStatus } from "./liveConnectionStates.js";
import { createEmptyVideoStreams } from "./liveMediaCleanup.js";

export function createEmptyLiveSession() {
  return {
    attemptId: null,
    events: [],
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

export function findTrackSlot(event, fallbackIndex) {
  const raw = [
    event.track?.id,
    event.track?.label,
    ...(event.streams ?? []).flatMap((stream) => [stream.id, stream.label])
  ].filter(Boolean).join(" ").toLowerCase();
  if (event.track?.kind === "audio" || raw.includes("audio")) {
    return "audio";
  }
  if (raw.includes("track.video_2")) {
    return "thermal";
  }
  if (raw.includes("track.video_1")) {
    return "rgb";
  }
  return fallbackIndex === 0 ? "rgb" : "thermal";
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
    return `미디어 ${mediaCount} / 데이터 ${dataChannelCount}`;
  }
  if (mediaCount > 0) {
    return `미디어 ${mediaCount}`;
  }
  if (dataChannelCount > 0) {
    return `데이터 ${dataChannelCount}`;
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
