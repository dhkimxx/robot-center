import {
  formatDateTime,
  formatElapsedTime,
  makeLiveStatusLabel,
  makeStatusLabel,
  makeStatusTone
} from "../../utils/formatters.js";
import {
  createEmptyLiveSession,
  formatMediaChannelCount,
  formatStreamingSubscriberCount
} from "./liveHelpers.js";

const browserMediaSlots = [
  ["rgb", "RGB"],
  ["thermal", "Thermal"],
  ["audio", "음성"]
];

export function getRobotLiveStatusSummary({ liveSessions = {}, target }) {
  if (!target) {
    return {
      channelLabel: "",
      connectionLabel: "연결 대기",
      recordingLabel: "녹화 대기",
      recordingState: { isActive: false, label: "녹화 대기", tone: "idle" },
      session: createEmptyLiveSession(),
      streamLabel: "송출 대기"
    };
  }
  const session = liveSessions[target.key] ?? createEmptyLiveSession();
  const recordingState = makeRecordingStateFromLiveStatus(target.liveStatus?.recording);
  const connectionLabel = target.liveStatus?.connection
    ? makeConnectionLabelFromLiveStatus(target.liveStatus.connection)
    : makeLiveStatusLabel(session.status);
  const streamSource = target.liveStatus?.stream;
  return {
    channelLabel: target.isStreaming
      ? `${formatMediaChannelCount(streamSource)} / ${formatStreamingSubscriberCount(streamSource)}`
      : "",
    connectionLabel,
    recordingLabel: recordingState.label,
    recordingState,
    session,
    streamLabel: makeStreamLabelFromLiveStatus(target.liveStatus?.stream, target)
  };
}

export function makeLiveRobotDiagnostics({ now = Date.now(), session, target } = {}) {
  const currentSession = session ?? createEmptyLiveSession();
  if (!target) {
    return [
      {
        detail: "관제할 로봇을 선택해야 합니다.",
        key: "target",
        label: "선택 로봇",
        tone: "waiting",
        value: "선택 대기"
      }
    ];
  }

  const stream = target.liveStatus?.stream;
  return [
    makeOperatorConnectionDiagnostic(currentSession),
    makePublisherMediaDiagnostic(stream, now),
    makePublisherDataDiagnostic(stream, now),
    makeRecordingDiagnostic(target.liveStatus?.recording)
  ];
}

export function makeStreamLabelFromLiveStatus(stream, target) {
  switch (stream?.state) {
    case "streaming":
      return "송출 중";
    case "ended":
      return "임무 종료";
    case "waiting":
      return "송출 대기";
    default:
      return target?.isStreaming ? "송출 중" : target?.liveLabel ?? "송출 대기";
  }
}

export function makeConnectionLabelFromLiveStatus(connection) {
  switch (connection?.state) {
    case "online":
      return "연결됨";
    case "fault":
      return "장애";
    case "disconnected":
      return "연결 끊김";
    case "offline":
      return "오프라인";
    default:
      return "연결 대기";
  }
}

export function makeRecordingStateFromLiveStatus(recording) {
  switch (recording?.state) {
    case "recording":
      return { isActive: true, label: "녹화 중", tone: "recording" };
    case "uploaded":
      return { isActive: false, label: "저장 완료", tone: "available" };
    case "failed":
      return { isActive: false, label: "녹화 오류", tone: "danger" };
    case "stale":
      return { isActive: false, label: "녹화 확인 필요", tone: "idle" };
    case "unknown":
      return { isActive: false, label: "녹화 확인 중", tone: "idle" };
    case "idle":
    default:
      return { isActive: false, label: "녹화 대기", tone: "idle" };
  }
}

function makeOperatorConnectionDiagnostic(session) {
  const receivedSlots = makeBrowserMediaReceiptLabel(session);
  return {
    detail: receivedSlots || "브라우저 수신 대기",
    key: "operator",
    label: "관제 연결",
    tone: makeDiagnosticToneFromLiveSession(session?.status),
    value: makeLiveStatusLabel(session?.status ?? "disconnected")
  };
}

function makePublisherMediaDiagnostic(stream, now) {
  const trackCount = normalizeCount(stream?.trackCount);
  const lastMediaAt = stream?.lastMediaAt ?? stream?.lastTrackAt;
  const details = [];
  if (lastMediaAt) {
    details.push(`최근 영상/음성 ${formatElapsedTime(lastMediaAt, now)}`);
  }
  if (trackCount > 0) {
    details.push(`송출 채널 ${trackCount}개`);
  }
  if ((stream?.diagnostics?.reconnectCount ?? 0) > 0) {
    details.push(`재접속 ${stream.diagnostics.reconnectCount}회`);
  }
  return {
    detail: details.join(" · ") || "로봇 영상/음성 대기",
    key: "media",
    label: "로봇 영상/음성",
    tone: stream?.state === "streaming" && trackCount > 0 ? "ok" : stream?.state === "fault" ? "danger" : "waiting",
    value: trackCount > 0 ? "송출 중" : "송출 대기"
  };
}

function makePublisherDataDiagnostic(stream, now) {
  const dataChannelCount = normalizeCount(stream?.dataChannelCount);
  const lastDataAt = stream?.lastDataAt;
  const details = [];
  if (lastDataAt) {
    details.push(`최근 수신 ${formatElapsedTime(lastDataAt, now)}`);
  }
  if (dataChannelCount > 0) {
    details.push(`수신 채널 ${dataChannelCount}개`);
  }
  return {
    detail: details.join(" · ") || "센서/이벤트 수신 대기",
    key: "data",
    label: "센서/이벤트 데이터",
    tone: dataChannelCount > 0 && lastDataAt ? "ok" : "waiting",
    value: lastDataAt ? "수신 중" : dataChannelCount > 0 ? "연결됨" : "수신 대기"
  };
}

function makeRecordingDiagnostic(recording) {
  const recordingState = makeRecordingStateFromLiveStatus(recording);
  const chunk = recording?.latestChunk;
  return {
    detail: chunk ? makeRecordingChunkDetail(chunk, recording) : recording?.reason || "녹화 구간 대기",
    key: "recording",
    label: "녹화",
    tone: makeDiagnosticToneFromRecordingState(recordingState.tone),
    value: recordingState.label
  };
}

function makeRecordingChunkDetail(chunk, recording) {
  const chunkLabel = chunk.chunkIndex || chunk.chunkIndex === 0 ? `구간 #${chunk.chunkIndex}` : "녹화 구간";
  const statusLabel = makeStatusLabel(chunk.status ?? recording?.latestChunkStatus ?? recording?.state);
  return `${chunkLabel} · ${formatDateTime(chunk.startedAt)} - ${formatDateTime(chunk.endedAt)} · ${statusLabel}`;
}

function makeBrowserMediaReceiptLabel(session) {
  const labels = browserMediaSlots
    .filter(([slot]) => Boolean(session?.videoStreams?.[slot]))
    .map(([, label]) => label);
  return labels.length > 0 ? `관제 화면 수신 ${labels.join(", ")}` : "";
}

function makeDiagnosticToneFromLiveSession(status) {
  const tone = makeStatusTone(status);
  if (tone === "ok" || tone === "danger") {
    return tone;
  }
  return "waiting";
}

function makeDiagnosticToneFromRecordingState(tone) {
  if (tone === "danger") {
    return "danger";
  }
  if (tone === "recording" || tone === "available") {
    return "ok";
  }
  return "waiting";
}

function normalizeCount(value) {
  const numberValue = Number(value);
  return Number.isFinite(numberValue) && numberValue > 0 ? numberValue : 0;
}
