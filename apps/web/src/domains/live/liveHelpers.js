import { makeMissionRobotKey } from "../missions/missionHelpers.js";

export function createEmptyLiveSession() {
  return {
    events: [],
    sensor: null,
    status: "disconnected",
    telemetry: null,
    videoStreams: { rgb: null, thermal: null, audio: null }
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
  if (raw.includes("thermal")) {
    return "thermal";
  }
  if (raw.includes("rgb")) {
    return "rgb";
  }
  return fallbackIndex === 0 ? "rgb" : "thermal";
}

export function findRobotCodeFromDataMessage(message) {
  try {
    const parsed = JSON.parse(message);
    return parsed?.robotCode
      ?? parsed?.payload?.robotCode
      ?? parsed?.rawPayload?.robotCode
      ?? parsed?.rawPayload?.payload?.robotCode
      ?? "";
  } catch {
    return "";
  }
}

export function getSampleRobotCode(sample) {
  return sample?.robotCode
    ?? sample?.payload?.robotCode
    ?? sample?.rawPayload?.robotCode
    ?? sample?.rawPayload?.payload?.robotCode
    ?? "";
}

export function findLatestSampleForRobot(samples, robotCode) {
  if (!Array.isArray(samples) || samples.length === 0) {
    return null;
  }
  if (!robotCode) {
    return samples[0] ?? null;
  }
  return samples.find((sample) => getSampleRobotCode(sample) === robotCode) ?? null;
}

export function formatMediaChannelCount(streamingStatus) {
  const channelCount = streamingStatus?.publishedTracks?.length ?? 0;
  return channelCount > 0 ? `${channelCount}개 채널` : "송출 대기";
}

export function getStreamingSubscriberCount(streamingStatus) {
  return streamingStatus?.subscriberCount
    ?? streamingStatus?.operatorCount
    ?? streamingStatus?.browserCount
    ?? streamingStatus?.viewerCount
    ?? null;
}

export function formatStreamingSubscriberCount(streamingStatus) {
  const subscriberCount = getStreamingSubscriberCount(streamingStatus);
  return subscriberCount === null ? "관제 -" : `관제 ${subscriberCount}명`;
}
