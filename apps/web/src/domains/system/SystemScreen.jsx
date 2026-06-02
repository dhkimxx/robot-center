import { useState } from "react";
import { RiDeleteBin6Line } from "react-icons/ri";
import { componentLabels } from "../../config/controlCenterConfig.js";
import ConfirmDialog from "../../components/ConfirmDialog.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import Button from "../../components/ui/Button.jsx";
import ListRow from "../../components/ui/ListRow.jsx";
import MetricStrip from "../../components/ui/MetricStrip.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton, PanelSkeleton, SkeletonBlock } from "../../components/ui/Skeleton.jsx";
import { cn } from "../../utils/cn.js";

export default function SystemScreen({ dataLoadState, onClearObjectStorage, onClearSensorData, statusError, systemStatus }) {
  const [clearConfirmOpen, setClearConfirmOpen] = useState(false);
  const [clearing, setClearing] = useState(false);
  const [sensorClearConfirmOpen, setSensorClearConfirmOpen] = useState(false);
  const [clearingSensors, setClearingSensors] = useState(false);
  const isInitialLoading = Boolean(dataLoadState?.isInitialLoading);
  const components = systemStatus?.components ?? [];
  const rooms = systemStatus?.sfuRooms ?? [];
  const environment = systemStatus?.config?.environment ?? "";
  const isProduction = environment === "production";
  const objectStorageUsage = normalizeObjectStorageUsage(systemStatus?.objectStorage);
  const summaryItems = [
    ["등록 로봇", systemStatus?.summary?.robots ?? 0],
    ["전체 임무", systemStatus?.summary?.missions ?? 0],
    ["녹화 항목", systemStatus?.summary?.recordings ?? 0],
    ["실시간 연결", systemStatus?.summary?.sfuRooms ?? rooms.length]
  ];

  async function confirmClearObjectStorage() {
    if (!onClearObjectStorage || clearing) {
      return;
    }
    setClearing(true);
    try {
      await onClearObjectStorage();
      setClearConfirmOpen(false);
    } finally {
      setClearing(false);
    }
  }

  async function confirmClearSensorData() {
    if (!onClearSensorData || clearingSensors) {
      return;
    }
    setClearingSensors(true);
    try {
      await onClearSensorData();
      setSensorClearConfirmOpen(false);
    } finally {
      setClearingSensors(false);
    }
  }

  return (
    <>
      <section className="grid h-full min-h-0 grid-cols-[400px_minmax(0,1fr)] gap-3 max-[980px]:grid-cols-1">
        <div className="grid min-h-0 content-start gap-3 overflow-auto">
          <Surface>
            <SectionHeader title="운영 요약" meta={isInitialLoading ? "확인 중" : statusError ? "응답 대기" : "정상 수신"} />
            {isInitialLoading ? (
              <div className="flex min-h-9 flex-wrap items-center gap-2 rounded-lg border border-slate-700/70 bg-slate-950/20 px-3">
                <SkeletonBlock className="h-6 w-24 rounded-full" />
                <SkeletonBlock className="h-6 w-24 rounded-full" />
                <SkeletonBlock className="h-6 w-24 rounded-full" />
              </div>
            ) : (
              <MetricStrip items={summaryItems.map(([label, value]) => ({ label, value }))} />
            )}
          </Surface>

          <Surface>
            <SectionHeader title="서비스 상태" meta={isInitialLoading ? "확인 중" : `${components.length}개 항목`} />
            {isInitialLoading ? (
              <ListSkeleton count={4} />
            ) : statusError ? (
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

          <Surface>
            <SectionHeader title="테스트 관리" meta={environment || "environment unknown"} />
            <div className="grid gap-3">
              {isInitialLoading ? <PanelSkeleton rows={3} /> : <ObjectStorageUsagePanel usage={objectStorageUsage} />}
              <div className="grid gap-3 rounded-lg border border-red-400/15 bg-red-400/[0.06] p-3">
                <div>
                  <strong className="block text-sm font-black text-red-100">Object Storage 전체 삭제</strong>
                  <span className="mt-1 block text-xs font-semibold leading-relaxed text-red-100/70">
                    MinIO bucket의 모든 object와 녹화 파일 availability metadata를 정리합니다.
                  </span>
                </div>
                <Button
                  className="justify-self-start"
                  disabled={isProduction || !onClearObjectStorage || clearing}
                  onClick={() => setClearConfirmOpen(true)}
                  variant="danger"
                >
                  <RiDeleteBin6Line aria-hidden="true" />
                  전체 삭제
                </Button>
              </div>
              <div className="grid gap-3 rounded-lg border border-red-400/15 bg-red-400/[0.06] p-3">
                <div>
                  <strong className="block text-sm font-black text-red-100">Sensor 데이터 전체 삭제</strong>
                  <span className="mt-1 block text-xs font-semibold leading-relaxed text-red-100/70">
                    저장된 sensor descriptor와 sample을 정리합니다. 새 telemetry가 들어오면 다시 생성됩니다.
                  </span>
                </div>
                <Button
                  className="justify-self-start"
                  disabled={isProduction || !onClearSensorData || clearingSensors}
                  onClick={() => setSensorClearConfirmOpen(true)}
                  variant="danger"
                >
                  <RiDeleteBin6Line aria-hidden="true" />
                  전체 삭제
                </Button>
              </div>
            </div>
          </Surface>
        </div>

        <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
          <SectionHeader title="실시간 연결" meta={isInitialLoading ? "확인 중" : `${rooms.length}개`} />
          <div className="grid min-h-0 auto-rows-max content-start gap-2 overflow-auto pr-1">
            {isInitialLoading ? (
              <ListSkeleton count={4} />
            ) : rooms.length === 0 ? (
              <EmptyState>연결된 세션이 없습니다.</EmptyState>
            ) : (
	              rooms.map((room) => {
	                const peerSummaries = createRoomPeerSummaries(room);
	                const robotCount = countRoomRobotPublishers(room);
	                const mediaCount = countRoomPublishedTracks(room);
	                const roomState = makeRoomStreamingState(room);
	                return (
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
	                      <StatusBadge tone={robotCount > 0 ? "success" : "neutral"}>로봇 {robotCount}대</StatusBadge>
	                      <StatusBadge tone={room.operatorCount > 0 ? "info" : "neutral"}>관제 {room.operatorCount}명</StatusBadge>
	                      <StatusBadge tone={room.recorderCount > 0 ? "warning" : "neutral"}>녹화 {room.recorderCount}개</StatusBadge>
	                    </div>
	                  </div>

	                  <div className="grid grid-cols-3 gap-2 max-[760px]:grid-cols-1">
	                    <RoomMetric label="미디어" value={`${mediaCount}개`} />
	                    <RoomMetric label="연결 주체" value={`${peerSummaries.length}개`} />
	                    <RoomMetric label="상태" value={roomState} />
	                  </div>

	                  <div className="flex flex-wrap gap-1.5">
	                    {peerSummaries.map((peer) => (
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
	                );
	              })
	            )}
          </div>
        </Surface>
      </section>
      {clearConfirmOpen ? (
        <ConfirmDialog
          cancelLabel="취소"
          confirmLabel={clearing ? "삭제 중" : "전체 삭제"}
          description="MinIO bucket의 모든 object를 삭제하고 녹화 파일 availability metadata를 초기화합니다. active recorder가 연결되어 있으면 이후 object가 다시 생성될 수 있습니다."
          onCancel={() => {
            if (!clearing) {
              setClearConfirmOpen(false);
            }
          }}
          onConfirm={confirmClearObjectStorage}
          subject={systemStatus?.config?.minioBucket ?? "object storage"}
          title="Object Storage 전체 삭제"
          tone="danger"
        />
      ) : null}
      {sensorClearConfirmOpen ? (
        <ConfirmDialog
          cancelLabel="취소"
          confirmLabel={clearingSensors ? "삭제 중" : "전체 삭제"}
          description="저장된 sensor descriptor와 sample을 모두 삭제합니다. active recorder가 telemetry를 다시 받으면 sensor 데이터가 다시 생성됩니다."
          onCancel={() => {
            if (!clearingSensors) {
              setSensorClearConfirmOpen(false);
            }
          }}
          onConfirm={confirmClearSensorData}
          subject="Sensor 데이터"
          title="Sensor 데이터 전체 삭제"
          tone="danger"
        />
      ) : null}
    </>
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

export function createRoomPeerSummaries(room) {
  const summaries = [];
  const seen = new Set();
  const addPeer = (peer) => {
    const key = makePeerSummaryKey(peer);
    if (!key || seen.has(key)) {
      return;
    }
    seen.add(key);
    summaries.push(peer);
  };

  (room?.publishers ?? []).forEach((publisher) => {
    if (publisher?.robotCode) {
      addPeer({
        peerId: publisher.publisherPeerId,
        role: "robot",
        robotCode: publisher.robotCode
      });
    }
  });
  (room?.peers ?? []).forEach((peer) => {
    if (peer?.role === "robot") {
      addPeer(peer);
      return;
    }
    if (peer?.role === "operator" || peer?.role === "recorder") {
      addPeer(peer);
    }
  });
  return summaries;
}

export function countRoomRobotPublishers(room) {
  const publisherRobotCodes = new Set(
    (room?.publishers ?? [])
      .map((publisher) => publisher?.robotCode)
      .filter(Boolean)
  );
  if (publisherRobotCodes.size > 0) {
    return publisherRobotCodes.size;
  }
  return new Set(
    (room?.peers ?? [])
      .filter((peer) => peer?.role === "robot" && peer.robotCode)
      .map((peer) => peer.robotCode)
  ).size;
}

export function countRoomPublishedTracks(room) {
  const publisherTrackCount = (room?.publishers ?? []).reduce((sum, publisher) => (
    sum + Math.max(0, Number(publisher?.trackCount) || 0)
  ), 0);
  if (publisherTrackCount > 0) {
    return publisherTrackCount;
  }
  return room?.publishedTracks?.length ?? 0;
}

export function makeRoomStreamingState(room) {
  const isPublishing = (room?.publishers ?? []).some((publisher) => (
    publisher?.state === "publishing" && Math.max(0, Number(publisher?.trackCount) || 0) > 0
  ));
  if (isPublishing) {
    return "송출 중";
  }
  if ((room?.publishers ?? []).length > 0 || countRoomRobotPublishers(room) > 0) {
    return "연결됨";
  }
  return "대기";
}

function makePeerSummaryKey(peer) {
  if (!peer?.role) {
    return "";
  }
  if (peer.role === "robot") {
    return peer.robotCode ? `robot:${peer.robotCode}` : `robot-peer:${peer.peerId}`;
  }
  if (peer.role === "operator") {
    return peer.selectedRobotCode ? `operator:${peer.selectedRobotCode}` : `operator-peer:${peer.peerId}`;
  }
  if (peer.role === "recorder") {
    return "recorder";
  }
  return `${peer.role}:${peer.peerId ?? ""}`;
}

function ObjectStorageUsagePanel({ usage }) {
  if (!usage || usage.status !== "ok" || usage.totalBytes <= 0) {
    return (
      <div className="rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
        <strong className="block text-sm font-black text-slate-100">Object Storage 용량</strong>
        <span className="mt-1 block text-xs font-semibold leading-relaxed text-slate-500">
          용량 정보를 불러오지 못했습니다.
        </span>
      </div>
    );
  }

  const percent = clampStoragePercent(usage.usedPercent);
  const chartColor = makeStorageChartColor(percent);

  return (
    <div className="grid gap-3 rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
      <div className="grid grid-cols-[96px_minmax(0,1fr)] items-center gap-3 max-[420px]:grid-cols-1">
        <div
          aria-label={`Object storage ${formatStoragePercent(percent)} used`}
          className="relative h-24 w-24 rounded-full"
          role="img"
          style={{
            background: `conic-gradient(${chartColor} ${percent}%, rgba(148, 163, 184, 0.20) ${percent}% 100%)`
          }}
        >
          <div className="absolute inset-3 grid place-items-center rounded-full border border-slate-500/15 bg-command-950 text-center">
            <div>
              <strong className="block text-lg font-black text-slate-50">{formatStoragePercent(percent)}</strong>
              <span className="block text-[10px] font-bold text-slate-500">사용 중</span>
            </div>
          </div>
        </div>
        <div className="min-w-0">
          <strong className="block truncate text-sm font-black text-slate-50">{usage.bucket || "Object Storage"}</strong>
          <span className="mt-1 block text-xs font-semibold text-slate-500">저장 가능 용량 대비 사용률</span>
          <strong className="mt-2 block text-sm font-black text-slate-100">
            {formatStorageByteCount(usage.usedBytes)} / {formatStorageByteCount(usage.totalBytes)}
          </strong>
          <span className="mt-1 block text-xs font-semibold text-emerald-200/80">
            가용 {formatStorageByteCount(usage.availableBytes)}
          </span>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-2">
        <StorageMetric label="Bucket 사용량" value={formatStorageByteCount(usage.bucketUsedBytes)} />
        <StorageMetric label="Object 수" value={`${usage.objectCount.toLocaleString()}개`} />
      </div>
    </div>
  );
}

function StorageMetric({ label, value }) {
  return (
    <div className="min-w-0 rounded-lg border border-slate-500/20 bg-white/[0.035] px-3 py-2">
      <span className="block truncate text-xs font-bold text-slate-500">{label}</span>
      <strong className="mt-1 block truncate text-sm font-black text-slate-100">{value}</strong>
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

function normalizeObjectStorageUsage(rawUsage) {
  if (!rawUsage) {
    return null;
  }
  const totalBytes = readStorageNumber(rawUsage.totalBytes);
  const usedBytes = readStorageNumber(rawUsage.usedBytes);
  const rawPercent = readStorageNumber(rawUsage.usedPercent);
  const calculatedPercent = totalBytes > 0 ? (usedBytes / totalBytes) * 100 : 0;
  return {
    availableBytes: readStorageNumber(rawUsage.availableBytes),
    bucket: rawUsage.bucket ?? "",
    bucketUsedBytes: readStorageNumber(rawUsage.bucketUsedBytes),
    objectCount: Math.max(0, Math.round(readStorageNumber(rawUsage.objectCount))),
    status: rawUsage.status ?? "unavailable",
    totalBytes,
    usedBytes,
    usedPercent: Number.isFinite(rawPercent) && rawPercent > 0 ? rawPercent : calculatedPercent
  };
}

function readStorageNumber(value) {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return 0;
  }
  return numberValue;
}

function clampStoragePercent(value) {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return 0;
  }
  return Math.min(100, Math.max(0, numberValue));
}

function formatStoragePercent(value) {
  const percent = clampStoragePercent(value);
  if (percent === 0 || percent >= 10) {
    return `${percent.toFixed(0)}%`;
  }
  return `${percent.toFixed(1)}%`;
}

function formatStorageByteCount(value) {
  const bytes = Math.max(0, readStorageNumber(value));
  if (bytes === 0) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let amount = bytes;
  let unitIndex = 0;
  while (amount >= 1024 && unitIndex < units.length - 1) {
    amount /= 1024;
    unitIndex += 1;
  }
  const digits = unitIndex === 0 || amount >= 100 ? 0 : amount >= 10 ? 1 : 2;
  return `${amount.toFixed(digits)} ${units[unitIndex]}`;
}

function makeStorageChartColor(percent) {
  if (percent >= 90) {
    return "#fb7185";
  }
  if (percent >= 75) {
    return "#fbbf24";
  }
  return "#38bdf8";
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
