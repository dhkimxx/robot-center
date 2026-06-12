import { useCallback, useRef } from "react";
import Button from "../../../components/ui/Button.jsx";
import { cn } from "../../../utils/cn.js";
import {
  liveDashboardGridColumns,
  liveDashboardWidgetSpecs
} from "../liveDashboardLayout.js";

const dashboardRowEstimate = 9;

export function LiveDashboardGrid({
  isEditing = false,
  layout,
  onMoveWidget,
  onResizeWidget,
  renderWidget
}) {
  const gridRef = useRef(null);

  const startPointerResize = useCallback((event, widget) => {
    if (!isEditing || !gridRef.current) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();

    const startX = event.clientX;
    const startY = event.clientY;
    const startW = widget.w;
    const startH = widget.h;
    const gridRect = gridRef.current.getBoundingClientRect();
    const columnWidth = Math.max(1, gridRect.width / liveDashboardGridColumns);
    const rowHeight = Math.max(42, gridRect.height / dashboardRowEstimate);

    function handlePointerMove(pointerMoveEvent) {
      const nextW = startW + Math.round((pointerMoveEvent.clientX - startX) / columnWidth);
      const nextH = startH + Math.round((pointerMoveEvent.clientY - startY) / rowHeight);
      onResizeWidget(widget.id, { h: nextH, w: nextW });
    }

    function handlePointerUp() {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("pointercancel", handlePointerUp);
    }

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", handlePointerUp);
    window.addEventListener("pointercancel", handlePointerUp);
  }, [isEditing, onResizeWidget]);

  return (
    <div
      className={cn(
        "grid h-full min-h-0 min-w-0 auto-rows-[minmax(42px,1fr)] grid-flow-row-dense grid-cols-12 gap-3 overflow-hidden",
        "max-[1180px]:flex max-[1180px]:flex-col max-[1180px]:overflow-auto"
      )}
      data-testid="live-dashboard-grid"
      ref={gridRef}
    >
      {layout.widgets.map((widget) => (
        <DashboardWidgetFrame
          isEditing={isEditing}
          key={widget.id}
          onMoveBackward={() => onMoveWidget(widget.id, "backward")}
          onMoveForward={() => onMoveWidget(widget.id, "forward")}
          onResize={(nextSize) => onResizeWidget(widget.id, nextSize)}
          onResizePointerDown={(event) => startPointerResize(event, widget)}
          widget={widget}
        >
          {renderWidget(widget.id)}
        </DashboardWidgetFrame>
      ))}
    </div>
  );
}

function DashboardWidgetFrame({
  children,
  isEditing,
  onMoveBackward,
  onMoveForward,
  onResize,
  onResizePointerDown,
  widget
}) {
  const spec = liveDashboardWidgetSpecs[widget.id];
  const widgetLabel = spec?.label ?? widget.id;
  const widgetStyle = {
    gridColumn: `span ${widget.w}`,
    gridRow: `span ${widget.h}`
  };

  return (
    <div
      className={cn(
        "group relative min-h-[140px] min-w-0 overflow-hidden rounded-xl",
        isEditing && "outline outline-1 outline-sapphire-400/35"
      )}
      data-testid={`live-dashboard-widget-${widget.id}`}
      style={widgetStyle}
    >
      {children}
      {isEditing ? (
        <div className="absolute inset-0 z-20 rounded-xl border border-sapphire-300/30 bg-command-950/10 pointer-events-none">
          <div className="pointer-events-auto absolute left-2 top-2 flex items-center gap-1 rounded-lg border border-slate-500/25 bg-command-950/90 px-2 py-1 shadow-xl shadow-black/25">
            <span className="mr-1 text-[11px] font-black text-slate-200">{widgetLabel}</span>
            <button
              className="rounded border border-slate-500/20 px-1.5 py-0.5 text-[11px] font-black text-slate-300 hover:border-sapphire-300/40 hover:text-white"
              onClick={onMoveBackward}
              title="앞으로 이동"
              type="button"
            >
              ←
            </button>
            <button
              className="rounded border border-slate-500/20 px-1.5 py-0.5 text-[11px] font-black text-slate-300 hover:border-sapphire-300/40 hover:text-white"
              onClick={onMoveForward}
              title="뒤로 이동"
              type="button"
            >
              →
            </button>
          </div>
          <div className="pointer-events-auto absolute right-2 top-2 flex items-center gap-1 rounded-lg border border-slate-500/25 bg-command-950/90 p-1 shadow-xl shadow-black/25">
            <SizeButton label="폭 축소" onClick={() => onResize({ w: widget.w - 1 })}>W-</SizeButton>
            <SizeButton label="폭 확대" onClick={() => onResize({ w: widget.w + 1 })}>W+</SizeButton>
            <SizeButton label="높이 축소" onClick={() => onResize({ h: widget.h - 1 })}>H-</SizeButton>
            <SizeButton label="높이 확대" onClick={() => onResize({ h: widget.h + 1 })}>H+</SizeButton>
          </div>
          <button
            aria-label={`${widgetLabel} 위젯 크기 조절`}
            className="pointer-events-auto absolute bottom-1.5 right-1.5 h-6 w-6 cursor-nwse-resize rounded-md border border-sapphire-300/45 bg-sapphire-500/25 text-[11px] font-black text-sapphire-100 shadow-xl shadow-black/30"
            onPointerDown={onResizePointerDown}
            title="드래그해서 크기 조절"
            type="button"
          >
            ↘
          </button>
        </div>
      ) : null}
    </div>
  );
}

function SizeButton({ children, label, onClick }) {
  return (
    <Button
      aria-label={label}
      className="h-6 rounded-md px-1.5 text-[10px]"
      onClick={onClick}
      size="sm"
      title={label}
      variant="ghost"
    >
      {children}
    </Button>
  );
}
