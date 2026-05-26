import EmptyState from "../../../components/ui/EmptyState.jsx";
import { cn } from "../../../utils/cn.js";
import {
  makeLiveStatusLabel
} from "../../../utils/formatters.js";
import {
  createEmptyLiveSession,
  formatMediaChannelCount,
  formatStreamingSubscriberCount
} from "../liveHelpers.js";

export function MissionRobotSelector({
  liveSessions,
  missionTargets,
  selectedTarget,
  setSelectedMissionTargetKey
}) {
  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-2">
      {missionTargets.length === 0 ? (
        <EmptyState>임무에 배정된 로봇이 없습니다.</EmptyState>
      ) : (
        missionTargets.map((target) => {
          const session = liveSessions[target.key] ?? createEmptyLiveSession();
          return (
            <button
              className={cn(
                "flex min-h-[72px] items-start justify-between gap-3 rounded-lg border border-slate-500/20 bg-white/[0.045] p-3 text-left transition hover:border-sapphire-500/[0.34] hover:bg-sapphire-500/[0.09] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500",
                selectedTarget?.key === target.key && "border-sapphire-500/50 bg-sapphire-500/[0.09] shadow-[inset_3px_0_0_var(--color-sapphire)]"
              )}
              key={target.key}
              type="button"
              onClick={() => setSelectedMissionTargetKey(target.key)}
            >
              <div className="grid min-w-0 gap-1">
                <strong className="truncate text-sm font-bold text-slate-50">{target.robot?.displayName ?? target.robotCode}</strong>
                <span className="truncate text-xs font-semibold text-slate-400">{target.robotCode} / {target.liveLabel}</span>
                <span className="truncate text-xs font-semibold text-slate-400">
                  {target.isStreaming ? `${formatMediaChannelCount(target.streamingStatus)} / ${formatStreamingSubscriberCount(target.streamingStatus)}` : target.liveLabel}
                </span>
              </div>
              <small className="shrink-0 rounded-full bg-sapphire-500/[0.13] px-2 py-1 text-xs font-bold text-blue-100">
                {makeLiveStatusLabel(session.status)}
              </small>
            </button>
          );
        })
      )}
    </div>
  );
}
