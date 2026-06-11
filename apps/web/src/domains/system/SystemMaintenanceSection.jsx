import { RiDeleteBin6Line } from "react-icons/ri";
import Button from "../../components/ui/Button.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { PanelSkeleton } from "../../components/ui/Skeleton.jsx";
import { DatabaseUsagePanel, ObjectStorageUsagePanel, RecorderRuntimeUsagePanel } from "./SystemUsagePanels.jsx";

export default function SystemMaintenanceSection({
  canClearEventData,
  canClearObjectStorage,
  canClearRecorderRuntime,
  canClearSensorData,
  clearingEvents,
  clearingObjectStorage,
  clearingRecorderRuntime,
  clearingSensors,
  databaseUsage,
  environment,
  isInitialLoading,
  isProduction,
  objectStorageUsage,
  objectStorageDisabledReason,
  onRequestClearEventData,
  onRequestClearObjectStorage,
  onRequestClearRecorderRuntime,
  onRequestClearSensorData,
  recorderRuntimeDisabledReason,
  recorderRuntimeStatus
}) {
  return (
    <Surface>
      <SectionHeader title="테스트 관리" meta={environment || "environment unknown"} />
      <div className="grid gap-3">
        {isInitialLoading ? <PanelSkeleton rows={3} /> : <ObjectStorageUsagePanel usage={objectStorageUsage} />}
        {isInitialLoading ? <PanelSkeleton rows={3} /> : <DatabaseUsagePanel usage={databaseUsage} />}
        {isInitialLoading ? <PanelSkeleton rows={3} /> : <RecorderRuntimeUsagePanel status={recorderRuntimeStatus} />}
        <DangerActionPanel
          busy={clearingObjectStorage}
          description={objectStorageDisabledReason || "객체 스토리지의 모든 녹화 파일과 파일 상태 정보를 정리합니다."}
          disabled={Boolean(objectStorageDisabledReason) || !canClearObjectStorage || clearingObjectStorage}
          onClick={onRequestClearObjectStorage}
          title="객체 스토리지 전체 삭제"
        />
        <DangerActionPanel
          busy={clearingSensors}
          description="저장된 센서 정의와 센서값을 정리합니다. 새 telemetry가 들어오면 다시 생성됩니다."
          disabled={isProduction || !canClearSensorData || clearingSensors}
          onClick={onRequestClearSensorData}
          title="센서 데이터 전체 삭제"
        />
        <DangerActionPanel
          busy={clearingEvents}
          description="저장된 임무 이벤트와 객체 탐지 데이터를 정리합니다. 새 event가 들어오면 다시 생성됩니다."
          disabled={isProduction || !canClearEventData || clearingEvents}
          onClick={onRequestClearEventData}
          title="이벤트 데이터 전체 삭제"
        />
        <DangerActionPanel
          busy={clearingRecorderRuntime}
          description={recorderRuntimeDisabledReason || "녹화 서비스의 로컬 임시 파일을 정리합니다."}
          disabled={Boolean(recorderRuntimeDisabledReason) || !canClearRecorderRuntime || clearingRecorderRuntime}
          onClick={onRequestClearRecorderRuntime}
          title="녹화 런타임 파일 전체 삭제"
        />
      </div>
    </Surface>
  );
}

function DangerActionPanel({ busy, description, disabled, onClick, title }) {
  return (
    <div className="grid gap-3 rounded-lg border border-red-400/15 bg-red-400/[0.06] p-3">
      <div>
        <strong className="block text-sm font-black text-red-100">{title}</strong>
        <span className="mt-1 block text-xs font-semibold leading-relaxed text-red-100/70">
          {description}
        </span>
      </div>
      <Button
        className="justify-self-start"
        disabled={disabled}
        onClick={onClick}
        variant="danger"
      >
        <RiDeleteBin6Line aria-hidden="true" />
        {busy ? "삭제 중" : "전체 삭제"}
      </Button>
    </div>
  );
}
