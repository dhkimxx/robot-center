export const liveDashboardLayoutStorageKey = "robot-center.liveDashboardLayout";
export const liveDashboardLayoutVersion = 2;
export const liveDashboardGridColumns = 12;

export const liveDashboardWidgetSpecs = {
  event: {
    label: "이벤트",
    maxH: 9,
    maxW: 6,
    minH: 3,
    minW: 3,
    required: true
  },
  map: {
    label: "지도",
    maxH: 6,
    maxW: 8,
    minH: 2,
    minW: 3,
    required: false
  },
  rgb: {
    label: "RGB",
    maxH: 9,
    maxW: 12,
    minH: 4,
    minW: 4,
    required: true
  },
  sensor: {
    label: "센서",
    maxH: 6,
    maxW: 8,
    minH: 2,
    minW: 3,
    required: false
  },
  thermal: {
    label: "Thermal",
    maxH: 6,
    maxW: 8,
    minH: 2,
    minW: 3,
    required: false
  }
};

export const liveDashboardPresets = {
  classic: {
    label: "클래식",
    widgets: [
      { h: 5, id: "rgb", w: 6 },
      { h: 5, id: "thermal", w: 6 },
      { h: 3, id: "map", w: 6 },
      { h: 3, id: "sensor", w: 6 },
      { h: 3, id: "event", w: 12 }
    ]
  },
  cockpit: {
    label: "콕핏",
    widgets: [
      { h: 6, id: "rgb", w: 8 },
      { h: 6, id: "event", w: 4 },
      { h: 3, id: "thermal", w: 4 },
      { h: 3, id: "map", w: 4 },
      { h: 3, id: "sensor", w: 4 }
    ]
  },
  sensorFocus: {
    label: "센서 중심",
    widgets: [
      { h: 5, id: "rgb", w: 7 },
      { h: 5, id: "sensor", w: 5 },
      { h: 4, id: "event", w: 4 },
      { h: 4, id: "thermal", w: 4 },
      { h: 4, id: "map", w: 4 }
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

  return {
    presetId,
    version: liveDashboardLayoutVersion,
    widgets: widgetIds
      .map((widgetId) => {
        const fallback = preset.widgets.find((widget) => widget.id === widgetId)
          ?? { h: liveDashboardWidgetSpecs[widgetId].minH, id: widgetId, w: liveDashboardWidgetSpecs[widgetId].minW };
        return normalizeLiveDashboardWidget({
          ...fallback,
          ...incomingById.get(widgetId),
          id: widgetId
        });
      })
      .filter(Boolean)
  };
}

export function resizeLiveDashboardWidget(layout, widgetId, nextSize) {
  const normalized = normalizeLiveDashboardLayout(layout);
  return {
    ...normalized,
    widgets: normalized.widgets.map((widget) => (
      widget.id === widgetId
        ? normalizeLiveDashboardWidget({ ...widget, ...nextSize })
        : widget
    ))
  };
}

export function moveLiveDashboardWidget(layout, widgetId, direction) {
  const normalized = normalizeLiveDashboardLayout(layout);
  const widgets = [...normalized.widgets];
  const index = widgets.findIndex((widget) => widget.id === widgetId);
  if (index < 0) {
    return normalized;
  }
  const nextIndex = direction === "backward" ? index - 1 : index + 1;
  if (nextIndex < 0 || nextIndex >= widgets.length) {
    return normalized;
  }
  const [widget] = widgets.splice(index, 1);
  widgets.splice(nextIndex, 0, widget);
  return { ...normalized, widgets };
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

function normalizeLiveDashboardWidget(widget) {
  const spec = liveDashboardWidgetSpecs[widget?.id];
  if (!spec) {
    return null;
  }
  return {
    h: clampInteger(widget.h, spec.minH, spec.maxH, spec.minH),
    id: widget.id,
    w: clampInteger(widget.w, spec.minW, Math.min(spec.maxW, liveDashboardGridColumns), spec.minW)
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
