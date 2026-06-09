import StatusBadge from "../../components/ui/StatusBadge.jsx";
import {
  formatInteger,
  formatStorageByteCount,
  formatStoragePercent,
  makeRecorderRuntimeBlockingLabel,
  makeStorageChartColor
} from "./systemViewModel.js";

export function ObjectStorageUsagePanel({ usage }) {
  if (!usage || usage.status !== "ok" || usage.totalBytes <= 0) {
    return (
      <div className="rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
        <strong className="block text-sm font-black text-slate-100">객체 스토리지 용량</strong>
        <span className="mt-1 block text-xs font-semibold leading-relaxed text-slate-500">
          용량 정보를 불러오지 못했습니다.
        </span>
      </div>
    );
  }

  const percent = Math.min(100, Math.max(0, usage.usedPercent));
  const chartColor = makeStorageChartColor(percent);

  return (
    <div className="grid gap-3 rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
      <div className="grid grid-cols-[96px_minmax(0,1fr)] items-center gap-3 max-[420px]:grid-cols-1">
        <div
          aria-label={`객체 스토리지 ${formatStoragePercent(percent)} 사용 중`}
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
          <strong className="block truncate text-sm font-black text-slate-50">객체 스토리지</strong>
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
        <StorageMetric label="저장 파일 사용량" value={formatStorageByteCount(usage.bucketUsedBytes)} />
        <StorageMetric label="파일 수" value={`${usage.objectCount.toLocaleString()}개`} />
      </div>
    </div>
  );
}

export function DatabaseUsagePanel({ usage }) {
  if (!usage || usage.status !== "ok") {
    return (
      <div className="rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
        <strong className="block text-sm font-black text-slate-100">데이터베이스 용량</strong>
        <span className="mt-1 block text-xs font-semibold leading-relaxed text-slate-500">
          데이터베이스 용량 정보를 불러오지 못했습니다.
        </span>
      </div>
    );
  }

  const categories = usage.categories ?? [];
  return (
    <div className="grid gap-3 rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
      <div className="min-w-0">
        <strong className="block truncate text-sm font-black text-slate-50">데이터베이스</strong>
        <span className="mt-1 block text-xs font-semibold text-slate-500">현재 저장 사용량</span>
      </div>
      <div className="grid grid-cols-2 gap-2">
        <StorageMetric label="전체 크기" value={formatStorageByteCount(usage.databaseSizeBytes)} />
        <StorageMetric label="관리 테이블 크기" value={formatStorageByteCount(usage.trackedTableBytes)} />
      </div>
      {categories.length > 0 ? (
        <div className="grid gap-1.5">
          {categories.map((category) => (
            <div className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-2 rounded-lg border border-slate-500/15 bg-white/[0.025] px-3 py-2" key={category.id}>
              <div className="min-w-0">
                <strong className="block truncate text-xs font-black text-slate-200">{category.label}</strong>
                <span className="mt-0.5 block truncate text-[11px] font-bold text-slate-500">관련 항목 {formatInteger(category.tableCount)}개</span>
              </div>
              <div className="text-right">
                <strong className="block text-xs font-black text-slate-100">{formatStorageByteCount(category.totalBytes)}</strong>
                <span className="mt-0.5 block text-[11px] font-bold text-slate-500">추정 {formatInteger(category.rowCount)}건</span>
              </div>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}

export function RecorderRuntimeUsagePanel({ status }) {
  if (!status || status.status !== "ok") {
    return (
      <div className="rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
        <strong className="block text-sm font-black text-slate-100">녹화 런타임 용량</strong>
        <span className="mt-1 block text-xs font-semibold leading-relaxed text-slate-500">
          녹화 임시 파일 용량 정보를 불러오지 못했습니다.
        </span>
      </div>
    );
  }

  const percent = Math.min(100, Math.max(0, status.usedPercent));
  return (
    <div className="grid gap-3 rounded-lg border border-slate-500/20 bg-command-900/50 p-3">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <div className="min-w-0">
          <strong className="block truncate text-sm font-black text-slate-50">녹화 런타임 용량</strong>
          <span className="mt-1 block text-xs font-semibold text-slate-500">녹화 처리 중 사용하는 로컬 임시 파일</span>
        </div>
        <StatusBadge size="xs" tone={status.clearable ? "success" : "warning"}>
          {status.clearable ? "정리 가능" : "정리 대기"}
        </StatusBadge>
      </div>
      <div className="h-2 overflow-hidden rounded-full bg-slate-700/50">
        <div
          className="h-full rounded-full bg-sapphire-400"
          style={{ width: `${percent}%` }}
        />
      </div>
      <div className="grid grid-cols-2 gap-2">
        <StorageMetric label="임시 파일 사용량" value={formatStorageByteCount(status.usedBytes)} />
        <StorageMetric label="디스크 대비 사용률" value={formatStoragePercent(percent)} />
        <StorageMetric label="파일 수" value={`${formatInteger(status.files)}개`} />
        <StorageMetric label="청크 디렉터리" value={`${formatInteger(status.recordingDirectories)}개`} />
      </div>
      {!status.clearable && status.blockingReason ? (
        <span className="text-xs font-semibold text-amber-100/80">{makeRecorderRuntimeBlockingLabel(status.blockingReason)}</span>
      ) : null}
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
