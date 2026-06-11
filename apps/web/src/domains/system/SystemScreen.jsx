import { useState } from "react";
import { componentLabels } from "../../config/controlCenterConfig.js";
import ConfirmDialog from "../../components/ConfirmDialog.jsx";
import EmptyState from "../../components/ui/EmptyState.jsx";
import ListRow from "../../components/ui/ListRow.jsx";
import MetricStrip from "../../components/ui/MetricStrip.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton, SkeletonBlock } from "../../components/ui/Skeleton.jsx";
import SystemMaintenanceSection from "./SystemMaintenanceSection.jsx";
import SystemRealtimeConnections from "./SystemRealtimeConnections.jsx";
import {
  makeRecorderRuntimeDisabledReason,
  makeObjectStorageDisabledReason,
  makeSystemStatusLabel,
  makeSystemStatusTone,
  normalizeDatabaseUsage,
  normalizeObjectStorageUsage,
  normalizeRecorderRuntimeStatus
} from "./systemViewModel.js";

export default function SystemScreen({ dataLoadState, onClearEventData, onClearObjectStorage, onClearRecorderRuntime, onClearSensorData, statusError, systemStatus }) {
  const [clearConfirmOpen, setClearConfirmOpen] = useState(false);
  const [clearing, setClearing] = useState(false);
  const [sensorClearConfirmOpen, setSensorClearConfirmOpen] = useState(false);
  const [clearingSensors, setClearingSensors] = useState(false);
  const [eventClearConfirmOpen, setEventClearConfirmOpen] = useState(false);
  const [clearingEvents, setClearingEvents] = useState(false);
  const [recorderRuntimeClearConfirmOpen, setRecorderRuntimeClearConfirmOpen] = useState(false);
  const [clearingRecorderRuntime, setClearingRecorderRuntime] = useState(false);
  const isInitialLoading = Boolean(dataLoadState?.isInitialLoading);
  const components = systemStatus?.components ?? [];
  const rooms = systemStatus?.sfuRooms ?? [];
  const environment = systemStatus?.config?.environment ?? "";
  const isProduction = environment === "production";
  const objectStorageUsage = normalizeObjectStorageUsage(systemStatus?.objectStorage);
  const databaseUsage = normalizeDatabaseUsage(systemStatus?.database);
  const recorderRuntimeStatus = normalizeRecorderRuntimeStatus(systemStatus?.recorderRuntime);
  const objectStorageDisabledReason = makeObjectStorageDisabledReason({ isProduction, recorderRuntimeStatus });
  const recorderRuntimeDisabledReason = makeRecorderRuntimeDisabledReason({ isProduction, recorderRuntimeStatus });
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

  async function confirmClearEventData() {
    if (!onClearEventData || clearingEvents) {
      return;
    }
    setClearingEvents(true);
    try {
      await onClearEventData();
      setEventClearConfirmOpen(false);
    } finally {
      setClearingEvents(false);
    }
  }

  async function confirmClearRecorderRuntime() {
    if (!onClearRecorderRuntime || clearingRecorderRuntime) {
      return;
    }
    setClearingRecorderRuntime(true);
    try {
      await onClearRecorderRuntime();
      setRecorderRuntimeClearConfirmOpen(false);
    } finally {
      setClearingRecorderRuntime(false);
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

          <SystemMaintenanceSection
            canClearEventData={Boolean(onClearEventData)}
            canClearObjectStorage={Boolean(onClearObjectStorage)}
            canClearRecorderRuntime={Boolean(onClearRecorderRuntime)}
            canClearSensorData={Boolean(onClearSensorData)}
            clearingEvents={clearingEvents}
            clearingObjectStorage={clearing}
            clearingRecorderRuntime={clearingRecorderRuntime}
            clearingSensors={clearingSensors}
            databaseUsage={databaseUsage}
            environment={environment}
            isInitialLoading={isInitialLoading}
            isProduction={isProduction}
            objectStorageUsage={objectStorageUsage}
            objectStorageDisabledReason={objectStorageDisabledReason}
            onRequestClearEventData={() => setEventClearConfirmOpen(true)}
            onRequestClearObjectStorage={() => setClearConfirmOpen(true)}
            onRequestClearRecorderRuntime={() => setRecorderRuntimeClearConfirmOpen(true)}
            onRequestClearSensorData={() => setSensorClearConfirmOpen(true)}
            recorderRuntimeDisabledReason={recorderRuntimeDisabledReason}
            recorderRuntimeStatus={recorderRuntimeStatus}
          />
        </div>

        <SystemRealtimeConnections isInitialLoading={isInitialLoading} rooms={rooms} />
      </section>
      {clearConfirmOpen ? (
        <ConfirmDialog
          cancelLabel="취소"
          confirmLabel={clearing ? "삭제 중" : "전체 삭제"}
          description="객체 스토리지의 모든 파일을 삭제하고 녹화 파일 상태 정보를 초기화합니다. 진행 중인 녹화가 있으면 이후 파일이 다시 생성될 수 있습니다."
          onCancel={() => {
            if (!clearing) {
              setClearConfirmOpen(false);
            }
          }}
          onConfirm={confirmClearObjectStorage}
          subject="객체 스토리지"
          title="객체 스토리지 전체 삭제"
          tone="danger"
        />
      ) : null}
      {sensorClearConfirmOpen ? (
        <ConfirmDialog
          cancelLabel="취소"
          confirmLabel={clearingSensors ? "삭제 중" : "전체 삭제"}
          description="저장된 센서 정의와 센서값을 모두 삭제합니다. 진행 중인 녹화가 telemetry를 다시 받으면 센서 데이터가 다시 생성됩니다."
          onCancel={() => {
            if (!clearingSensors) {
              setSensorClearConfirmOpen(false);
            }
          }}
          onConfirm={confirmClearSensorData}
          subject="센서 데이터"
          title="센서 데이터 전체 삭제"
          tone="danger"
        />
      ) : null}
      {eventClearConfirmOpen ? (
        <ConfirmDialog
          cancelLabel="취소"
          confirmLabel={clearingEvents ? "삭제 중" : "전체 삭제"}
          description="저장된 mission.event와 detection.object 이벤트를 모두 삭제합니다. 진행 중인 녹화가 event를 다시 받으면 데이터가 다시 생성됩니다."
          onCancel={() => {
            if (!clearingEvents) {
              setEventClearConfirmOpen(false);
            }
          }}
          onConfirm={confirmClearEventData}
          subject="이벤트 데이터"
          title="이벤트 데이터 전체 삭제"
          tone="danger"
        />
      ) : null}
      {recorderRuntimeClearConfirmOpen ? (
        <ConfirmDialog
          cancelLabel="취소"
          confirmLabel={clearingRecorderRuntime ? "삭제 중" : "전체 삭제"}
          description="녹화 서비스의 로컬 임시 파일을 모두 삭제합니다. 진행 중인 녹화 작업이 있으면 서버가 거부합니다."
          onCancel={() => {
            if (!clearingRecorderRuntime) {
              setRecorderRuntimeClearConfirmOpen(false);
            }
          }}
          onConfirm={confirmClearRecorderRuntime}
          subject="녹화 런타임 파일"
          title="녹화 런타임 파일 전체 삭제"
          tone="danger"
        />
      ) : null}
    </>
  );
}
