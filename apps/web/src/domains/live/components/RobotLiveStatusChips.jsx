import { cn } from "../../../utils/cn.js";
import { makeLiveStatusLabel } from "../../../utils/formatters.js";
import { createEmptyLiveSession, formatMediaChannelCount, formatStreamingSubscriberCount } from "../liveHelpers.js";

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

export function RobotLiveStatusChips({ className = "", summary, target }) {
  if (!target || !summary) {
    return null;
  }
  return (
    <div className={cn("flex min-w-0 flex-wrap items-center gap-1.5", className)}>
      <StateChip active={target.isStreaming} tone={target.isStreaming ? "streaming" : "idle"}>
        {summary.streamLabel}
      </StateChip>
      <StateChip active={summary.recordingState.isActive} tone={summary.recordingState.tone}>
        {summary.recordingLabel}
      </StateChip>
      <StateChip tone={summary.connectionLabel === "연결됨" ? "connected" : summary.connectionLabel === "장애" ? "danger" : "idle"}>
        {summary.connectionLabel}
      </StateChip>
      {summary.channelLabel ? (
        <span className="truncate text-xs font-semibold text-slate-500">{summary.channelLabel}</span>
      ) : null}
    </div>
  );
}

function makeStreamLabelFromLiveStatus(stream, target) {
  switch (stream?.state) {
    case "streaming":
      return "송출 중";
    case "ended":
      return "임무 종료";
    case "waiting":
      return "송출 대기";
    default:
      return target.isStreaming ? "송출 중" : target.liveLabel;
  }
}

function makeConnectionLabelFromLiveStatus(connection) {
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

function makeRecordingStateFromLiveStatus(recording) {
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

function StateChip({ active = false, children, tone }) {
  return (
    <span
      className={cn(
        "inline-flex h-6 items-center gap-1.5 rounded-full border px-2 text-[11px] font-bold",
        tone === "streaming" && "border-emerald-400/25 bg-emerald-400/[0.10] text-emerald-100",
        tone === "recording" && "border-amber-300/25 bg-amber-300/[0.10] text-amber-100",
        tone === "available" && "border-sapphire-400/25 bg-sapphire-400/[0.10] text-sapphire-100",
        tone === "connected" && "border-sapphire-400/25 bg-sapphire-400/[0.10] text-sapphire-100",
        tone === "danger" && "border-red-400/25 bg-red-400/[0.10] text-red-100",
        tone === "idle" && "border-slate-500/20 bg-white/[0.04] text-slate-400"
      )}
    >
      <span
        aria-hidden
        className={cn(
          "h-1.5 w-1.5 rounded-full",
          tone === "streaming" && "bg-emerald-300",
          tone === "recording" && "bg-amber-200",
          (tone === "available" || tone === "connected") && "bg-sapphire-300",
          tone === "danger" && "bg-red-300",
          tone === "idle" && "bg-slate-500",
          active && "motion-safe:animate-pulse"
        )}
      />
      {children}
    </span>
  );
}
