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
  createSystemClearActions,
  makeSystemStatusLabel,
  makeSystemStatusTone,
  normalizeDatabaseUsage,
  normalizeObjectStorageUsage,
  normalizeRecorderRuntimeStatus
} from "./systemViewModel.js";

export default function SystemScreen({ dataLoadState, onClearEventData, onClearObjectStorage, onClearRecorderRuntime, onClearSensorData, statusError, systemStatus }) {
  const [activeClearActionID, setActiveClearActionID] = useState("");
  const [clearingActionID, setClearingActionID] = useState("");
  const isInitialLoading = Boolean(dataLoadState?.isInitialLoading);
  const components = systemStatus?.components ?? [];
  const rooms = systemStatus?.sfuRooms ?? [];
  const environment = systemStatus?.config?.environment ?? "";
  const isProduction = environment === "production";
  const objectStorageUsage = normalizeObjectStorageUsage(systemStatus?.objectStorage);
  const databaseUsage = normalizeDatabaseUsage(systemStatus?.database);
  const recorderRuntimeStatus = normalizeRecorderRuntimeStatus(systemStatus?.recorderRuntime);
  const clearActionHandlers = {
    eventData: onClearEventData,
    objectStorage: onClearObjectStorage,
    recorderRuntime: onClearRecorderRuntime,
    sensorData: onClearSensorData
  };
  const clearActions = createSystemClearActions({
    canClearEventData: Boolean(onClearEventData),
    canClearObjectStorage: Boolean(onClearObjectStorage),
    canClearRecorderRuntime: Boolean(onClearRecorderRuntime),
    canClearSensorData: Boolean(onClearSensorData),
    clearingActionID,
    databaseUsage,
    isProduction,
    objectStorageUsage,
    recorderRuntimeStatus,
    statusReady: !isInitialLoading && !statusError && Boolean(systemStatus)
  });
  const activeClearAction = clearActions.find((action) => action.id === activeClearActionID);
  const summaryItems = [
    ["등록 로봇", systemStatus?.summary?.robots ?? 0],
    ["전체 임무", systemStatus?.summary?.missions ?? 0],
    ["녹화 항목", systemStatus?.summary?.recordings ?? 0],
    ["실시간 연결", systemStatus?.summary?.sfuRooms ?? rooms.length]
  ];

  async function confirmClearAction() {
    if (!activeClearAction || activeClearAction.disabled || clearingActionID) {
      return;
    }
    const clearActionHandler = clearActionHandlers[activeClearAction.id];
    if (!clearActionHandler) {
      return;
    }
    setClearingActionID(activeClearAction.id);
    try {
      await clearActionHandler();
      setActiveClearActionID("");
    } finally {
      setClearingActionID("");
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
            clearActions={clearActions}
            databaseUsage={databaseUsage}
            environment={environment}
            isInitialLoading={isInitialLoading}
            objectStorageUsage={objectStorageUsage}
            onRequestClearAction={setActiveClearActionID}
            recorderRuntimeStatus={recorderRuntimeStatus}
          />
        </div>

        <SystemRealtimeConnections isInitialLoading={isInitialLoading} rooms={rooms} />
      </section>
      {activeClearAction ? (
        <ConfirmDialog
          cancelLabel="취소"
          confirmLabel={activeClearAction.busy ? "삭제 중" : "전체 삭제"}
          description={activeClearAction.description}
          details={activeClearAction.targetMetrics}
          onCancel={() => {
            if (!clearingActionID) {
              setActiveClearActionID("");
            }
          }}
          onConfirm={confirmClearAction}
          subject={activeClearAction.subject}
          title={activeClearAction.title}
          tone="danger"
          warning={activeClearAction.impact}
        />
      ) : null}
    </>
  );
}
