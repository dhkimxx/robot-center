import { RiDeleteBin6Line } from "react-icons/ri";
import Button from "../../components/ui/Button.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { PanelSkeleton } from "../../components/ui/Skeleton.jsx";

export default function SystemMaintenanceSection({
  clearActions,
  environment,
  isInitialLoading,
  onRequestClearAction
}) {
  return (
    <Surface>
      <SectionHeader title="데이터 정리" meta={environment || "environment unknown"} />
      <div className="grid grid-cols-2 gap-3 max-[900px]:grid-cols-1">
        {isInitialLoading ? <PanelSkeleton rows={3} /> : null}
        {(clearActions ?? []).map((action) => (
          <DangerActionPanel
            action={action}
            key={action.id}
            onClick={() => onRequestClearAction(action.id)}
          />
        ))}
      </div>
    </Surface>
  );
}

function DangerActionPanel({ action, onClick }) {
  const description = action.disabledReason || action.description;
  return (
    <div className="grid gap-3 rounded-lg border border-red-400/15 bg-red-400/[0.06] p-3">
      <div>
        <strong className="block text-sm font-black text-red-100">{action.title}</strong>
        <span className="mt-1 block text-xs font-semibold leading-relaxed text-red-100/70">
          {description}
        </span>
      </div>
      {action.targetMetrics?.length > 0 ? (
        <div className="grid grid-cols-3 gap-2 max-[520px]:grid-cols-1">
          {action.targetMetrics.map((metric) => (
            <div className="min-w-0 rounded-lg border border-red-200/10 bg-black/10 px-3 py-2" key={metric.label}>
              <span className="block truncate text-[11px] font-bold text-red-100/55">{metric.label}</span>
              <strong className="mt-1 block truncate text-xs font-black text-red-50">{metric.value}</strong>
            </div>
          ))}
        </div>
      ) : null}
      <Button
        className="justify-self-start"
        disabled={action.disabled}
        onClick={onClick}
        variant="danger"
      >
        <RiDeleteBin6Line aria-hidden="true" />
        {action.busy ? action.busyLabel : action.buttonLabel}
      </Button>
    </div>
  );
}
