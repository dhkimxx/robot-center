import { componentLabels } from "../../config/controlCenterConfig.js";
import EmptyState from "../../components/ui/EmptyState.jsx";
import ListRow from "../../components/ui/ListRow.jsx";
import MetricStrip from "../../components/ui/MetricStrip.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { cn } from "../../utils/cn.js";

export default function SystemScreen({ statusError, systemStatus }) {
  const components = systemStatus?.components ?? [];
  const rooms = systemStatus?.sfuRooms ?? [];
  const summaryItems = [
    ["등록 로봇", systemStatus?.summary?.robots ?? 0],
    ["전체 임무", systemStatus?.summary?.missions ?? 0],
    ["녹화 항목", systemStatus?.summary?.recordings ?? 0],
    ["실시간 연결", systemStatus?.summary?.sfuRooms ?? rooms.length]
  ];

  return (
    <section className="grid h-full min-h-0 grid-cols-[400px_minmax(0,1fr)] gap-3 max-[980px]:grid-cols-1">
      <div className="grid min-h-0 content-start gap-3 overflow-auto">
        <Surface>
          <SectionHeader title="운영 요약" meta={statusError ? "응답 대기" : "정상 수신"} />
          <MetricStrip items={summaryItems.map(([label, value]) => ({ label, value }))} />
        </Surface>

        <Surface>
          <SectionHeader title="서비스 상태" meta={`${components.length}개 항목`} />
          {statusError ? (
            <EmptyState>응답 대기: {statusError}</EmptyState>
          ) : (
            <ul className="grid gap-2">
              {components.map((component) => (
                <ListRow
                  as="li"
                  key={component.name}
                  right={<StatusBadge size="xs" tone={makeSystemStatusTone(component.status)}>{makeSystemStatusLabel(component.status)}</StatusBadge>}
                  title={componentLabels[component.name] ?? component.name}
                >
                </ListRow>
              ))}
            </ul>
          )}
        </Surface>
      </div>

      <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
        <SectionHeader title="실시간 연결" meta={`${rooms.length}개`} />
        <div className="grid min-h-0 auto-rows-max content-start gap-2 overflow-auto pr-1">
          {rooms.length === 0 ? (
            <EmptyState>연결된 세션이 없습니다.</EmptyState>
          ) : (
            rooms.map((room) => (
              <div
                className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-4"
                key={room.roomId}
              >
                <div className="flex min-w-0 items-start justify-between gap-3 max-[760px]:grid">
                  <div className="min-w-0">
                    <strong className="block truncate text-base font-black text-slate-50">{room.roomId}</strong>
                    <span className="mt-1 block text-xs font-semibold text-slate-500">실시간 관제 연결</span>
                  </div>
                  <div className="flex flex-wrap justify-end gap-1.5 max-[760px]:justify-start">
                    <StatusBadge tone="success">로봇 {room.robotCount}대</StatusBadge>
                    <StatusBadge tone={room.operatorCount > 0 ? "info" : "neutral"}>관제 {room.operatorCount}명</StatusBadge>
                    <StatusBadge tone={room.recorderCount > 0 ? "warning" : "neutral"}>녹화 {room.recorderCount}개</StatusBadge>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-2 max-[760px]:grid-cols-1">
                  <RoomMetric label="미디어" value={`${room.publishedTracks?.length ?? 0}개`} />
                  <RoomMetric label="연결 주체" value={`${room.peers?.length ?? 0}개`} />
                  <RoomMetric label="상태" value={room.robotCount > 0 ? "송출 중" : "대기"} />
                </div>

                <div className="flex flex-wrap gap-1.5">
                  {(room.peers ?? []).map((peer) => (
                    <span
                      className={cn(
                        "inline-flex min-h-7 max-w-full items-center rounded-full border px-2.5 text-xs font-semibold",
                        peer.role === "robot" && "border-emerald-400/25 bg-emerald-400/[0.10] text-emerald-100",
                        peer.role === "operator" && "border-sapphire-400/25 bg-sapphire-400/[0.10] text-sapphire-100",
                        peer.role === "recorder" && "border-amber-300/25 bg-amber-300/[0.10] text-amber-100"
                      )}
                      key={peer.peerId}
                    >
                      <span className="truncate">{makePeerLabel(peer)}</span>
                    </span>
                  ))}
                </div>
              </div>
            ))
          )}
        </div>
      </Surface>
    </section>
  );
}

function RoomMetric({ label, value }) {
  return (
    <div className="rounded-lg border border-slate-500/20 bg-command-900/50 px-3 py-2">
      <span className="block text-xs font-bold text-slate-500">{label}</span>
      <strong className="mt-1 block text-sm font-black text-slate-100">{value}</strong>
    </div>
  );
}

function makeSystemStatusLabel(status) {
  const labels = {
    configured: "설정됨",
    degraded: "점검 필요",
    error: "오류",
    failed: "실패",
    ok: "정상",
    ready: "준비"
  };
  return labels[status] ?? status;
}

function makeSystemStatusTone(status) {
  if (["ok", "ready", "configured"].includes(status)) {
    return "success";
  }
  if (["degraded", "warning"].includes(status)) {
    return "warning";
  }
  if (["error", "failed"].includes(status)) {
    return "danger";
  }
  return "neutral";
}

function makePeerLabel(peer) {
  if (peer.role === "robot") {
    return peer.robotCode ? `로봇 ${peer.robotCode}` : "로봇";
  }
  if (peer.role === "operator") {
    return peer.selectedRobotCode ? `관제 ${peer.selectedRobotCode}` : "관제";
  }
  if (peer.role === "recorder") {
    return "녹화";
  }
  return peer.role;
}
