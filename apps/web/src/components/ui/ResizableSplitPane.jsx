import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { cn } from "../../utils/cn.js";

const defaultMinimumRatio = 0.2;
const defaultMaximumRatio = 0.8;
const splitHandleWidth = 10;
const splitStorageVersion = 2;

export function getSplitRatioBounds(containerWidth, leftMinWidth, rightMinWidth) {
  const paneWidth = containerWidth - splitHandleWidth;
  if (!Number.isFinite(paneWidth) || paneWidth <= 0) {
    return { maxRatio: defaultMaximumRatio, minRatio: defaultMinimumRatio };
  }

  const leftMinRatio = leftMinWidth / paneWidth;
  const rightMaxRatio = 1 - (rightMinWidth / paneWidth);
  const minRatio = Math.max(defaultMinimumRatio, leftMinRatio);
  const maxRatio = Math.min(defaultMaximumRatio, rightMaxRatio);

  if (minRatio > maxRatio) {
    return { maxRatio: defaultMaximumRatio, minRatio: defaultMinimumRatio };
  }

  return { maxRatio, minRatio };
}

export function clampSplitRatio(value, {
  containerWidth = 0,
  fallbackRatio = 0.6,
  leftMinWidth = 360,
  rightMinWidth = 320
} = {}) {
  const ratio = Number.isFinite(value) ? value : fallbackRatio;
  const { maxRatio, minRatio } = getSplitRatioBounds(containerWidth, leftMinWidth, rightMinWidth);
  return Math.min(Math.max(ratio, minRatio), maxRatio);
}

export function parseStoredSplitRatio(rawValue, fallbackRatio = 0.6) {
  if (!rawValue) {
    return fallbackRatio;
  }

  try {
    const storedValue = JSON.parse(rawValue);
    if (storedValue?.version !== splitStorageVersion || !Number.isFinite(storedValue?.ratio)) {
      return fallbackRatio;
    }
    return clampSplitRatio(storedValue.ratio, { fallbackRatio });
  } catch {
    return fallbackRatio;
  }
}

export function stringifyStoredSplitRatio(ratio) {
  return JSON.stringify({
    ratio,
    version: splitStorageVersion
  });
}

export default function ResizableSplitPane({
  className,
  defaultLeftRatio = 0.6,
  left,
  leftMinWidth = 560,
  right,
  rightMinWidth = 360,
  storageKey,
  style,
  ...props
}) {
  const containerRef = useRef(null);
  const [containerWidth, setContainerWidth] = useState(0);
  const [leftRatio, setLeftRatio] = useState(() => (
    getInitialLeftRatio(storageKey, defaultLeftRatio)
  ));

  const splitBounds = useMemo(
    () => getSplitRatioBounds(containerWidth, leftMinWidth, rightMinWidth),
    [containerWidth, leftMinWidth, rightMinWidth]
  );

  const updateRatio = useCallback((nextRatio, nextContainerWidth = containerWidth) => {
    setLeftRatio(clampSplitRatio(nextRatio, {
      containerWidth: nextContainerWidth,
      fallbackRatio: defaultLeftRatio,
      leftMinWidth,
      rightMinWidth
    }));
  }, [containerWidth, defaultLeftRatio, leftMinWidth, rightMinWidth]);

  const updateRatioFromPointer = useCallback((clientX) => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const bounds = container.getBoundingClientRect();
    const nextRatio = (clientX - bounds.left) / (bounds.width - splitHandleWidth);
    updateRatio(nextRatio, bounds.width);
  }, [updateRatio]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return undefined;
    }

    function refreshWidth() {
      setContainerWidth(container.getBoundingClientRect().width);
    }

    refreshWidth();

    if (typeof ResizeObserver === "undefined") {
      window.addEventListener("resize", refreshWidth);
      return () => window.removeEventListener("resize", refreshWidth);
    }

    const resizeObserver = new ResizeObserver(refreshWidth);
    resizeObserver.observe(container);
    return () => resizeObserver.disconnect();
  }, []);

  useEffect(() => {
    updateRatio(leftRatio);
  }, [leftRatio, updateRatio]);

  useEffect(() => {
    if (!storageKey || typeof window === "undefined") {
      return;
    }
    window.localStorage.setItem(storageKey, stringifyStoredSplitRatio(leftRatio));
  }, [leftRatio, storageKey]);

  function handlePointerDown(event) {
    event.preventDefault();
    updateRatioFromPointer(event.clientX);

    const previousCursor = document.body.style.cursor;
    const previousUserSelect = document.body.style.userSelect;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    function handlePointerMove(pointerMoveEvent) {
      updateRatioFromPointer(pointerMoveEvent.clientX);
    }

    function handlePointerUp() {
      document.body.style.cursor = previousCursor;
      document.body.style.userSelect = previousUserSelect;
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("pointercancel", handlePointerUp);
    }

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", handlePointerUp);
    window.addEventListener("pointercancel", handlePointerUp);
  }

  function handleKeyDown(event) {
    const step = event.shiftKey ? 0.08 : 0.03;

    if (event.key === "ArrowLeft") {
      event.preventDefault();
      updateRatio(leftRatio - step);
      return;
    }

    if (event.key === "ArrowRight") {
      event.preventDefault();
      updateRatio(leftRatio + step);
      return;
    }

    if (event.key === "Home") {
      event.preventDefault();
      updateRatio(splitBounds.minRatio);
      return;
    }

    if (event.key === "End") {
      event.preventDefault();
      updateRatio(splitBounds.maxRatio);
      return;
    }

    if (event.key === "Enter") {
      event.preventDefault();
      updateRatio(defaultLeftRatio);
    }
  }

  const mergedStyle = {
    ...style,
    "--split-left-min": `${leftMinWidth}px`,
    "--split-left-size": `calc((100% - ${splitHandleWidth}px) * ${leftRatio})`,
    "--split-right-min": `${rightMinWidth}px`,
    "--split-right-size": `calc((100% - ${splitHandleWidth}px) * ${1 - leftRatio})`
  };

  return (
    <section
      className={cn(
        "grid h-full min-h-0 items-stretch gap-0 max-[1180px]:grid-cols-1 max-[1180px]:gap-3",
        "grid-cols-[minmax(var(--split-left-min),var(--split-left-size))_10px_minmax(var(--split-right-min),var(--split-right-size))]",
        className
      )}
      ref={containerRef}
      style={mergedStyle}
      {...props}
    >
      <div className="min-h-0 min-w-0">{left}</div>
      <div
        aria-label="목록과 상세 영역 너비 조절"
        aria-orientation="vertical"
        aria-valuemax={Math.round(splitBounds.maxRatio * 100)}
        aria-valuemin={Math.round(splitBounds.minRatio * 100)}
        aria-valuenow={Math.round(leftRatio * 100)}
        className="group flex min-h-0 cursor-col-resize items-stretch justify-center rounded-md focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500 max-[1180px]:hidden"
        role="separator"
        tabIndex={0}
        onDoubleClick={() => updateRatio(defaultLeftRatio)}
        onKeyDown={handleKeyDown}
        onPointerDown={handlePointerDown}
      >
        <span className="my-1 w-px rounded-full bg-slate-700/80 transition group-hover:bg-sapphire-300/80 group-focus-visible:bg-sapphire-300" />
      </div>
      <div className="min-h-0 min-w-0">{right}</div>
    </section>
  );
}

function getInitialLeftRatio(storageKey, defaultLeftRatio) {
  if (!storageKey || typeof window === "undefined") {
    return defaultLeftRatio;
  }

  return parseStoredSplitRatio(window.localStorage.getItem(storageKey), defaultLeftRatio);
}
