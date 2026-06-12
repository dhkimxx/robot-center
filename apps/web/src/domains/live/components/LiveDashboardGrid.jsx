import { useCallback, useEffect, useRef, useState } from "react";
import { cn } from "../../../utils/cn.js";
import {
  calculateLiveDashboardRowCount,
  liveDashboardGridColumns,
  liveDashboardWidgetSpecs
} from "../liveDashboardLayout.js";

const dashboardGapPx = 12;
const minimumDashboardRowHeightPx = 42;

export function LiveDashboardGrid({
  isEditing = false,
  layout,
  onMoveWidget,
  onResizeWidget,
  renderWidget
}) {
  const gridRef = useRef(null);
  const [activeEdit, setActiveEdit] = useState(null);
  const previewLayout = activeEdit
    ? {
        ...layout,
        widgets: layout.widgets.map((widget) => (
          widget.id === activeEdit.widgetId
            ? createPreviewWidget(widget, activeEdit)
            : widget
        ))
      }
    : layout;
  const rowCount = calculateLiveDashboardRowCount(previewLayout);
  const rowHeight = useDashboardRowHeight(gridRef, rowCount);

  useEffect(() => {
    if (!isEditing) {
      setActiveEdit(null);
    }
  }, [isEditing]);

  const startPointerEdit = useCallback((event, widget, action) => {
    if (!isEditing || !gridRef.current) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();

    const startX = event.clientX;
    const startY = event.clientY;
    const startWidget = { ...widget };
    const gridRect = gridRef.current.getBoundingClientRect();
    const columnStep = Math.max(1, gridRect.width / liveDashboardGridColumns);
    const rowStep = Math.max(1, rowHeight + dashboardGapPx);
    const pointerId = event.pointerId;
    const pointerTarget = event.currentTarget;
    let latestColumnDelta = 0;
    let latestRowDelta = 0;

    pointerTarget.setPointerCapture?.(pointerId);
    setActiveEdit({ action, columnDelta: 0, rowDelta: 0, widgetId: widget.id });

    function handlePointerMove(pointerMoveEvent) {
      latestColumnDelta = Math.round((pointerMoveEvent.clientX - startX) / columnStep);
      latestRowDelta = Math.round((pointerMoveEvent.clientY - startY) / rowStep);
      setActiveEdit({
        action,
        columnDelta: latestColumnDelta,
        rowDelta: latestRowDelta,
        widgetId: widget.id
      });
    }

    function handlePointerUp() {
      if (action === "move") {
        onMoveWidget(widget.id, {
          x: startWidget.x + latestColumnDelta,
          y: startWidget.y + latestRowDelta
        });
      } else {
        onResizeWidget(widget.id, {
          h: action.includes("s") ? startWidget.h + latestRowDelta : startWidget.h,
          w: action.includes("e") ? startWidget.w + latestColumnDelta : startWidget.w
        });
      }
      setActiveEdit(null);
      pointerTarget.releasePointerCapture?.(pointerId);
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("pointercancel", handlePointerUp);
    }

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", handlePointerUp);
    window.addEventListener("pointercancel", handlePointerUp);
  }, [isEditing, onMoveWidget, onResizeWidget, rowHeight]);

  return (
    <div
      className={cn(
        "grid h-full min-h-0 min-w-0 gap-3 overflow-auto rounded-xl",
        "max-[1180px]:flex max-[1180px]:flex-col"
      )}
      data-testid="live-dashboard-grid"
      ref={gridRef}
      style={{
        gridTemplateColumns: `repeat(${liveDashboardGridColumns}, minmax(0, 1fr))`,
        gridTemplateRows: `repeat(${rowCount}, ${rowHeight}px)`
      }}
    >
      {layout.widgets.map((widget) => (
        <DashboardWidgetFrame
          activeEdit={activeEdit?.widgetId === widget.id ? activeEdit : null}
          isEditing={isEditing}
          key={widget.id}
          onPointerEdit={(event, action) => startPointerEdit(event, widget, action)}
          widget={widget}
        >
          {renderWidget(widget.id)}
        </DashboardWidgetFrame>
      ))}
    </div>
  );
}

function DashboardWidgetFrame({
  activeEdit,
  children,
  isEditing,
  onPointerEdit,
  widget
}) {
  const spec = liveDashboardWidgetSpecs[widget.id];
  const widgetLabel = spec?.label ?? widget.id;
  const displayedWidget = activeEdit ? createPreviewWidget(widget, activeEdit) : widget;
  const widgetStyle = {
    gridColumn: `${displayedWidget.x + 1} / span ${displayedWidget.w}`,
    gridRow: `${displayedWidget.y + 1} / span ${displayedWidget.h}`
  };

  return (
    <div
      className={cn(
        "group relative min-h-[140px] min-w-0 overflow-hidden rounded-xl transition-[outline-color,box-shadow]",
        activeEdit && "z-30",
        isEditing && "outline outline-1 outline-sapphire-400/45 shadow-[0_0_0_1px_rgba(56,189,248,0.08)]"
      )}
      data-testid={`live-dashboard-widget-${widget.id}`}
      style={widgetStyle}
    >
      {children}
      {isEditing ? (
        <div className="absolute inset-0 z-20 rounded-xl border border-sapphire-300/25 bg-command-950/5 pointer-events-none">
          <button
            aria-label={`${widgetLabel} 패널 이동`}
            className="pointer-events-auto absolute left-2 top-2 flex h-8 cursor-move select-none items-center gap-2 rounded-lg border border-sapphire-300/35 bg-command-950/90 px-3 text-xs font-black text-slate-100 shadow-xl shadow-black/25 hover:border-sapphire-200/60"
            data-testid={`live-dashboard-drag-${widget.id}`}
            onPointerDown={(event) => onPointerEdit(event, "move")}
            title="드래그해서 패널 이동"
            type="button"
          >
            <span className="text-sapphire-200">⋮⋮</span>
            {widgetLabel}
          </button>
          <button
            aria-label={`${widgetLabel} 패널 폭 조절`}
            className="pointer-events-auto absolute bottom-8 right-0 top-8 w-2 cursor-ew-resize rounded-l bg-sapphire-300/0 transition hover:bg-sapphire-300/45"
            data-testid={`live-dashboard-resize-east-${widget.id}`}
            onPointerDown={(event) => onPointerEdit(event, "e")}
            title="드래그해서 폭 조절"
            type="button"
          />
          <button
            aria-label={`${widgetLabel} 패널 높이 조절`}
            className="pointer-events-auto absolute bottom-0 left-8 right-8 h-2 cursor-ns-resize rounded-t bg-sapphire-300/0 transition hover:bg-sapphire-300/45"
            data-testid={`live-dashboard-resize-south-${widget.id}`}
            onPointerDown={(event) => onPointerEdit(event, "s")}
            title="드래그해서 높이 조절"
            type="button"
          />
          <button
            aria-label={`${widgetLabel} 패널 크기 조절`}
            className="pointer-events-auto absolute bottom-1.5 right-1.5 h-5 w-5 cursor-nwse-resize rounded-md border border-sapphire-300/55 bg-command-950/90 text-[10px] font-black text-sapphire-100 shadow-xl shadow-black/30 hover:bg-sapphire-500/25"
            data-testid={`live-dashboard-resize-corner-${widget.id}`}
            onPointerDown={(event) => onPointerEdit(event, "se")}
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

function useDashboardRowHeight(gridRef, rowCount) {
  const [rowHeight, setRowHeight] = useState(minimumDashboardRowHeightPx);

  useEffect(() => {
    const gridElement = gridRef.current;
    if (!gridElement || typeof ResizeObserver !== "function") {
      return undefined;
    }

    const updateRowHeight = () => {
      const rect = gridElement.getBoundingClientRect();
      const availableHeight = rect.height - Math.max(0, rowCount - 1) * dashboardGapPx;
      const nextRowHeight = Math.max(minimumDashboardRowHeightPx, Math.floor(availableHeight / rowCount));
      setRowHeight((current) => current === nextRowHeight ? current : nextRowHeight);
    };

    updateRowHeight();
    const observer = new ResizeObserver(updateRowHeight);
    observer.observe(gridElement);
    return () => observer.disconnect();
  }, [gridRef, rowCount]);

  return rowHeight;
}

function createPreviewWidget(widget, activeEdit) {
  const spec = liveDashboardWidgetSpecs[widget.id];
  if (!spec) {
    return widget;
  }
  const nextX = activeEdit.action === "move"
    ? clampGridNumber(widget.x + activeEdit.columnDelta, 0, liveDashboardGridColumns - spec.minW)
    : widget.x;
  const nextY = activeEdit.action === "move"
    ? Math.max(0, widget.y + activeEdit.rowDelta)
    : widget.y;
  const maxWidth = Math.min(spec.maxW, liveDashboardGridColumns - nextX);
  const nextWidth = activeEdit.action.includes("e")
    ? clampGridNumber(widget.w + activeEdit.columnDelta, spec.minW, maxWidth)
    : Math.min(widget.w, maxWidth);
  const nextHeight = activeEdit.action.includes("s")
    ? clampGridNumber(widget.h + activeEdit.rowDelta, spec.minH, spec.maxH)
    : widget.h;

  return {
    ...widget,
    h: nextHeight,
    w: nextWidth,
    x: nextX,
    y: nextY
  };
}

function clampGridNumber(value, min, max) {
  return Math.min(max, Math.max(min, Math.round(value)));
}
