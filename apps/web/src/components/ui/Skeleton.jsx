import { cn } from "../../utils/cn.js";

export function SkeletonBlock({ className }) {
  return (
    <span
      aria-hidden="true"
      className={cn("block rounded-md bg-slate-700/45 motion-safe:animate-pulse", className)}
    />
  );
}

export function SkeletonText({ className, lines = 1 }) {
  return (
    <div className={cn("grid gap-2", className)} aria-hidden="true">
      {Array.from({ length: lines }).map((_, index) => (
        <SkeletonBlock
          className={cn("h-3", index === lines - 1 && lines > 1 ? "w-2/3" : "w-full")}
          key={index}
        />
      ))}
    </div>
  );
}

export function ListSkeleton({ count = 4, className }) {
  return (
    <div className={cn("grid gap-2", className)} aria-label="목록을 불러오는 중">
      {Array.from({ length: count }).map((_, index) => (
        <div
          className="grid min-h-[64px] grid-cols-[minmax(0,1fr)_72px] items-center gap-3 rounded-lg border border-slate-700/70 bg-white/[0.035] px-3 py-2"
          key={index}
        >
          <div className="grid gap-2">
            <SkeletonBlock className="h-4 w-3/5" />
            <SkeletonBlock className="h-3 w-4/5" />
            <SkeletonBlock className="h-3 w-2/5" />
          </div>
          <SkeletonBlock className="h-6 rounded-full" />
        </div>
      ))}
    </div>
  );
}

export function PanelSkeleton({ className, rows = 4 }) {
  return (
    <div
      aria-label="패널을 불러오는 중"
      className={cn("grid content-start gap-4 rounded-xl border border-slate-700/70 bg-white/[0.035] p-4", className)}
    >
      <div className="grid gap-2">
        <SkeletonBlock className="h-5 w-1/3" />
        <SkeletonBlock className="h-3 w-1/4" />
      </div>
      <div className="grid gap-3">
        {Array.from({ length: rows }).map((_, index) => (
          <SkeletonBlock className={cn("h-10", index % 2 === 0 ? "w-full" : "w-5/6")} key={index} />
        ))}
      </div>
    </div>
  );
}
