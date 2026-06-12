import Button from "../../../components/ui/Button.jsx";
import { cn } from "../../../utils/cn.js";
import { liveDashboardPresets } from "../liveDashboardLayout.js";

export function LiveDashboardControls({
  className = "",
  isEditing,
  layout,
  onCancel,
  onChangePreset,
  onReset,
  onSave,
  onStartEdit
}) {
  return (
    <div className={cn("flex flex-wrap items-center justify-end gap-2", className)} data-testid="live-dashboard-controls">
      <label className="inline-flex min-h-8 items-center gap-2 rounded-lg border border-slate-500/20 bg-command-950/55 px-2 text-xs font-bold text-slate-400">
        <span className="shrink-0">레이아웃</span>
        <select
          className="h-6 rounded-md border border-slate-600/60 bg-command-900 px-2 text-xs font-black text-slate-100 outline-none focus:border-sapphire-400"
          onChange={(event) => onChangePreset(event.target.value)}
          value={layout.presetId}
        >
          {Object.entries(liveDashboardPresets).map(([presetId, preset]) => (
            <option key={presetId} value={presetId}>
              {preset.label}
            </option>
          ))}
        </select>
      </label>

      {isEditing ? (
        <>
          <Button onClick={onSave} size="sm" variant="primary">
            저장
          </Button>
          <Button onClick={onCancel} size="sm" variant="secondary">
            취소
          </Button>
          <Button onClick={onReset} size="sm" variant="ghost">
            기본값
          </Button>
        </>
      ) : (
        <>
          <Button onClick={onStartEdit} size="sm" variant="secondary">
            레이아웃 편집
          </Button>
          <Button onClick={onReset} size="sm" variant="ghost">
            기본값
          </Button>
        </>
      )}
    </div>
  );
}
