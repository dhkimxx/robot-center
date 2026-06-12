export const liveDashboardLayoutStorageKey = "robot-center.liveDashboardLayout";
export const liveDashboardLayoutVersion = 3;
export const liveDashboardGridColumns = 24;
export const liveDashboardGridMinRows = 13;

export const liveDashboardWidgetSpecs = {
  event: {
    label: "이벤트",
    maxH: 14,
    maxW: 12,
    minH: 4,
    minW: 6,
    required: true
  },
  map: {
    label: "지도",
    maxH: 12,
    maxW: 18,
    minH: 3,
    minW: 6,
    required: false
  },
  rgb: {
    label: "RGB",
    maxH: 16,
    maxW: 24,
    minH: 5,
    minW: 8,
    required: true
  },
  sensor: {
    label: "센서",
    maxH: 12,
    maxW: 18,
    minH: 3,
    minW: 6,
    required: false
  },
  thermal: {
    label: "Thermal",
    maxH: 12,
    maxW: 18,
    minH: 3,
    minW: 6,
    required: false
  }
};

export const liveDashboardPresets = {
  classic: {
    label: "클래식",
    widgets: [
      { h: 8, id: "rgb", w: 12, x: 0, y: 0 },
      { h: 8, id: "thermal", w: 12, x: 12, y: 0 },
      { h: 5, id: "map", w: 12, x: 0, y: 8 },
      { h: 5, id: "sensor", w: 12, x: 12, y: 8 },
      { h: 4, id: "event", w: 24, x: 0, y: 13 }
    ]
  },
  cockpit: {
    label: "콕핏",
    widgets: [
      { h: 8, id: "rgb", w: 16, x: 0, y: 0 },
      { h: 8, id: "event", w: 8, x: 16, y: 0 },
      { h: 5, id: "thermal", w: 8, x: 0, y: 8 },
      { h: 5, id: "map", w: 8, x: 8, y: 8 },
      { h: 5, id: "sensor", w: 8, x: 16, y: 8 }
    ]
  },
  sensorFocus: {
    label: "센서 중심",
    widgets: [
      { h: 7, id: "rgb", w: 12, x: 0, y: 0 },
      { h: 7, id: "sensor", w: 12, x: 12, y: 0 },
      { h: 6, id: "event", w: 8, x: 0, y: 7 },
      { h: 6, id: "thermal", w: 8, x: 8, y: 7 },
      { h: 6, id: "map", w: 8, x: 16, y: 7 }
    ]
  }
};

export const defaultLiveDashboardPresetId = "cockpit";

export function createPresetLiveDashboardLayout(presetId = defaultLiveDashboardPresetId) {
  const resolvedPresetId = liveDashboardPresets[presetId] ? presetId : defaultLiveDashboardPresetId;
  const preset = liveDashboardPresets[resolvedPresetId];
  return normalizeLiveDashboardLayout({
    presetId: resolvedPresetId,
    version: liveDashboardLayoutVersion,
    widgets: preset.widgets
  });
}

export function normalizeLiveDashboardLayout(layout = {}) {
  const presetId = liveDashboardPresets[layout.presetId] ? layout.presetId : defaultLiveDashboardPresetId;
  const preset = liveDashboardPresets[presetId];
  const incomingWidgets = Array.isArray(layout.widgets) ? layout.widgets : preset.widgets;
  const incomingById = new Map(
    incomingWidgets
      .filter((widget) => liveDashboardWidgetSpecs[widget?.id])
      .map((widget) => [widget.id, widget])
  );
  const widgetIds = mergeWidgetIds(
    preset.widgets.map((widget) => widget.id),
    Object.keys(liveDashboardWidgetSpecs),
    incomingWidgets.map((widget) => widget?.id)
  );
  const widgets = widgetIds
    .map((widgetId) => {
      const fallback = preset.widgets.find((widget) => widget.id === widgetId)
        ?? createDefaultWidget(widgetId);
      return normalizeLiveDashboardWidget({
        ...fallback,
        ...incomingById.get(widgetId),
        id: widgetId
      }, fallback);
    })
    .filter(Boolean);

  return {
    presetId,
    version: liveDashboardLayoutVersion,
    widgets: compactLiveDashboardWidgets(resolveLiveDashboardCollisions(widgets))
  };
}

export function moveLiveDashboardWidget(layout, widgetId, nextPosition) {
  const normalized = normalizeLiveDashboardLayout(layout);
  const widgets = normalized.widgets.map((widget) => (
    widget.id === widgetId
      ? normalizeLiveDashboardWidget({ ...widget, ...nextPosition }, widget)
      : widget
  ));
  return {
    ...normalized,
    widgets: compactLiveDashboardWidgets(resolveLiveDashboardCollisions(widgets, widgetId), widgetId)
  };
}

export function resizeLiveDashboardWidget(layout, widgetId, nextSize) {
  const normalized = normalizeLiveDashboardLayout(layout);
  const widgets = normalized.widgets.map((widget) => (
    widget.id === widgetId
      ? normalizeLiveDashboardWidget(resizeLiveDashboardWidgetAtCurrentPosition(widget, nextSize), widget)
      : widget
  ));
  return {
    ...normalized,
    widgets: compactLiveDashboardWidgets(resolveLiveDashboardCollisions(widgets, widgetId), widgetId)
  };
}

export function resolveLiveDashboardCollisions(widgets, activeWidgetId = "") {
  const sortedWidgets = [...widgets].sort(compareWidgetsForLayout);
  const activeWidget = activeWidgetId
    ? sortedWidgets.find((widget) => widget.id === activeWidgetId)
    : null;
  const passiveWidgets = activeWidget
    ? sortedWidgets.filter((widget) => widget.id !== activeWidgetId)
    : sortedWidgets;
  const orderedWidgets = activeWidget ? [activeWidget, ...passiveWidgets] : passiveWidgets;
  const placedWidgets = [];

  orderedWidgets.forEach((widget) => {
    const nextWidget = { ...widget };
    let collision = findFirstCollision(nextWidget, placedWidgets);
    while (collision) {
      nextWidget.y = collision.y + collision.h;
      collision = findFirstCollision(nextWidget, placedWidgets);
    }
    placedWidgets.push(nextWidget);
  });

  return placedWidgets.sort(compareWidgetsForLayout);
}

export function compactLiveDashboardWidgets(widgets, lockedWidgetId = "") {
  const lockedWidget = lockedWidgetId
    ? widgets.find((widget) => widget.id === lockedWidgetId)
    : null;
  const placedWidgets = lockedWidget ? [{ ...lockedWidget }] : [];
  [...widgets].filter((widget) => widget.id !== lockedWidgetId).sort(compareWidgetsForLayout).forEach((widget) => {
    const nextWidget = { ...widget };
    while (nextWidget.y > 0 && !findFirstCollision({ ...nextWidget, y: nextWidget.y - 1 }, placedWidgets)) {
      nextWidget.y -= 1;
    }
    placedWidgets.push(nextWidget);
  });
  return placedWidgets.sort(compareWidgetsForLayout);
}

export function calculateLiveDashboardRowCount(layout) {
  const widgets = Array.isArray(layout?.widgets) ? layout.widgets : [];
  return Math.max(
    liveDashboardGridMinRows,
    ...widgets.map((widget) => Number(widget.y || 0) + Number(widget.h || 0))
  );
}

export function readLiveDashboardLayout(storage = getLiveDashboardStorage()) {
  if (!storage) {
    return createPresetLiveDashboardLayout();
  }
  try {
    const storedValue = JSON.parse(storage.getItem(liveDashboardLayoutStorageKey) || "{}");
    if (storedValue?.version !== liveDashboardLayoutVersion) {
      return createPresetLiveDashboardLayout();
    }
    return normalizeLiveDashboardLayout(storedValue);
  } catch {
    return createPresetLiveDashboardLayout();
  }
}

export function writeLiveDashboardLayout(layout, storage = getLiveDashboardStorage()) {
  const normalized = normalizeLiveDashboardLayout(layout);
  if (!storage) {
    return normalized;
  }
  try {
    storage.setItem(liveDashboardLayoutStorageKey, JSON.stringify(normalized));
  } catch {
    // Storage failures should not block live control.
  }
  return normalized;
}

function normalizeLiveDashboardWidget(widget, fallback = createDefaultWidget(widget?.id)) {
  const spec = liveDashboardWidgetSpecs[widget?.id];
  if (!spec) {
    return null;
  }
  const width = clampInteger(widget.w, spec.minW, Math.min(spec.maxW, liveDashboardGridColumns), fallback.w);
  return {
    h: clampInteger(widget.h, spec.minH, spec.maxH, fallback.h),
    id: widget.id,
    w: width,
    x: clampInteger(widget.x, 0, liveDashboardGridColumns - width, fallback.x),
    y: clampInteger(widget.y, 0, Number.MAX_SAFE_INTEGER, fallback.y)
  };
}

function resizeLiveDashboardWidgetAtCurrentPosition(widget, nextSize) {
  const spec = liveDashboardWidgetSpecs[widget.id];
  const nextWidth = Number(nextSize.w);
  if (!spec || !Number.isFinite(nextWidth)) {
    return { ...widget, ...nextSize };
  }
  const nextX = Number.isFinite(Number(nextSize.x)) ? Number(nextSize.x) : widget.x;
  return {
    ...widget,
    ...nextSize,
    w: Math.min(nextWidth, liveDashboardGridColumns - nextX)
  };
}

function createDefaultWidget(widgetId) {
  const spec = liveDashboardWidgetSpecs[widgetId];
  return {
    h: spec?.minH ?? 1,
    id: widgetId,
    w: spec?.minW ?? 1,
    x: 0,
    y: 0
  };
}

function mergeWidgetIds(presetIds, knownIds, incomingIds) {
  const ids = [];
  [...incomingIds, ...presetIds, ...knownIds].forEach((id) => {
    if (!liveDashboardWidgetSpecs[id] || ids.includes(id)) {
      return;
    }
    if (liveDashboardWidgetSpecs[id].required || presetIds.includes(id) || incomingIds.includes(id)) {
      ids.push(id);
    }
  });
  return ids;
}

function findFirstCollision(widget, widgets) {
  return widgets.find((otherWidget) => widgetsOverlap(widget, otherWidget));
}

function widgetsOverlap(left, right) {
  return left.x < right.x + right.w
    && left.x + left.w > right.x
    && left.y < right.y + right.h
    && left.y + left.h > right.y;
}

function compareWidgetsForLayout(left, right) {
  return left.y - right.y || left.x - right.x || left.id.localeCompare(right.id);
}

function clampInteger(value, min, max, fallback) {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return fallback;
  }
  return Math.min(max, Math.max(min, Math.round(numberValue)));
}

function getLiveDashboardStorage() {
  try {
    return globalThis.localStorage ?? null;
  } catch {
    return null;
  }
}
