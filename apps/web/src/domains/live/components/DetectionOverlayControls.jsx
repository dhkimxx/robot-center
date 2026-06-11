import { cn } from "../../../utils/cn.js";
import { detectionOverlaySettingLimits } from "../detectionOverlaySettings.js";

export function DetectionOverlayControls({ className = "", maxDetections, onChange, ttlSeconds }) {
  return (
    <div className={cn("flex flex-wrap items-center gap-2", className)}>
      <NumberSetting
        label="bbox 유지"
        max={detectionOverlaySettingLimits.ttlSeconds.max}
        min={detectionOverlaySettingLimits.ttlSeconds.min}
        onChange={(value) => onChange("ttlSeconds", value)}
        suffix="초"
        value={ttlSeconds}
      />
      <NumberSetting
        label="bbox 최대"
        max={detectionOverlaySettingLimits.maxDetections.max}
        min={detectionOverlaySettingLimits.maxDetections.min}
        onChange={(value) => onChange("maxDetections", value)}
        suffix="개"
        value={maxDetections}
      />
    </div>
  );
}

function NumberSetting({ label, max, min, onChange, suffix, value }) {
  return (
    <label className="inline-flex min-h-8 items-center gap-1 rounded-lg border border-slate-500/20 bg-command-950/55 px-1.5 text-xs font-bold text-slate-400">
      <span className="shrink-0">{label}</span>
      <input
        className="h-6 w-9 rounded-md border border-slate-600/60 bg-command-900 px-1 text-right text-xs font-black text-slate-100 outline-none focus:border-sapphire-400"
        max={max}
        min={min}
        onChange={(event) => onChange(event.target.value)}
        type="number"
        value={value}
      />
      <span className="shrink-0 text-slate-500">{suffix}</span>
    </label>
  );
}
