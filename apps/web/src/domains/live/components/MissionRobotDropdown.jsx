import { useEffect, useMemo, useRef, useState } from "react";
import { LuCheck, LuChevronDown } from "react-icons/lu";
import { cn } from "../../../utils/cn.js";
import { getRobotLiveStatusSummary, RobotLiveStatusChips } from "./RobotLiveStatusChips.jsx";

export function MissionRobotDropdown({
  liveSessions,
  missionTargets,
  selectedMissionTargetKey,
  setSelectedMissionTargetKey
}) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef(null);
  const selectedTarget = missionTargets.find((target) => target.key === selectedMissionTargetKey) ?? missionTargets[0] ?? null;
  const selectedSummary = useMemo(
    () => getRobotLiveStatusSummary({ liveSessions, target: selectedTarget }),
    [liveSessions, selectedTarget]
  );
  const disabled = missionTargets.length === 0;

  useEffect(() => {
    if (!open) {
      return undefined;
    }

    function closeOnOutsideClick(event) {
      if (rootRef.current && !rootRef.current.contains(event.target)) {
        setOpen(false);
      }
    }

    function closeOnEscape(event) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", closeOnOutsideClick);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("mousedown", closeOnOutsideClick);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  function selectTarget(targetKey) {
    setSelectedMissionTargetKey(targetKey);
    setOpen(false);
  }

  return (
    <div className="relative min-w-0" ref={rootRef}>
      <button
        aria-expanded={open}
        className={cn(
          "group flex h-12 w-full min-w-0 items-center justify-between gap-3 rounded-lg border px-3 text-left outline-none transition",
          "border-slate-500/20 bg-command-950/80 hover:border-sapphire-400/40 hover:bg-command-900/80 focus:border-sapphire-400 focus:ring-2 focus:ring-sapphire-500/20",
          disabled && "cursor-not-allowed opacity-60"
        )}
        disabled={disabled}
        type="button"
        onClick={() => setOpen((current) => !current)}
      >
        <SelectedRobotContent selectedSummary={selectedSummary} selectedTarget={selectedTarget} />
        <LuChevronDown
          className={cn(
            "h-5 w-5 shrink-0 text-slate-400 transition group-hover:text-slate-200",
            open && "rotate-180 text-slate-100"
          )}
          aria-hidden
        />
      </button>

      {open ? (
        <div className="absolute right-0 top-[calc(100%+8px)] z-30 w-[min(520px,calc(100vw-32px))] overflow-hidden rounded-lg border border-slate-500/20 bg-command-950 shadow-2xl shadow-black/40">
          <div className="border-b border-slate-500/10 px-3 py-2">
            <span className="text-xs font-bold text-slate-400">관제 로봇 선택</span>
          </div>
          <div className="max-h-80 overflow-y-auto p-1.5">
            {missionTargets.map((target) => {
              const active = target.key === selectedTarget?.key;
              const summary = getRobotLiveStatusSummary({ liveSessions, target });
              return (
                <button
                  className={cn(
                    "grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-md px-3 py-2.5 text-left transition",
                    active ? "bg-sapphire-500/15 text-slate-50" : "text-slate-200 hover:bg-white/[0.055]"
                  )}
                  key={target.key}
                  type="button"
                  onClick={() => selectTarget(target.key)}
                >
                  <RobotOptionContent summary={summary} target={target} />
                  <LuCheck className={cn("h-4 w-4 text-sapphire-200", !active && "opacity-0")} aria-hidden />
                </button>
              );
            })}
          </div>
        </div>
      ) : null}
    </div>
  );
}

function SelectedRobotContent({ selectedSummary, selectedTarget }) {
  if (!selectedTarget) {
    return (
      <span className="min-w-0 text-sm font-bold text-slate-400">
        선택 가능한 로봇 없음
      </span>
    );
  }

  return (
    <span className="grid min-w-0 flex-1 gap-0.5 overflow-hidden">
      <span className="flex min-w-0 items-baseline gap-2 overflow-hidden">
        <strong className="truncate text-sm font-extrabold text-slate-50">
          {selectedTarget.robot?.displayName ?? selectedTarget.robotCode}
        </strong>
        <span className="shrink-0 text-xs font-bold text-slate-500">{selectedTarget.robotCode}</span>
      </span>
      <span className="truncate text-xs font-bold text-slate-500">
        {selectedSummary.streamLabel} · {selectedSummary.recordingLabel} · {selectedSummary.connectionLabel}
      </span>
    </span>
  );
}

function RobotOptionContent({ summary, target }) {
  return (
    <span className="grid min-w-0 gap-1.5">
      <span className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5">
        <strong className="truncate text-sm font-bold text-slate-50">{target.robot?.displayName ?? target.robotCode}</strong>
        <span className="truncate text-xs font-bold text-slate-500">{target.robotCode}</span>
      </span>
      <RobotLiveStatusChips summary={summary} target={target} />
    </span>
  );
}
