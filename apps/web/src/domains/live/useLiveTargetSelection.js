import { useEffect, useMemo, useState } from "react";

const selectedLiveTargetStorageKey = "robot-center.selectedLiveTargetKey";

export function readSelectedLiveTargetKey() {
  try {
    return window.localStorage.getItem(selectedLiveTargetStorageKey) ?? "";
  } catch {
    return "";
  }
}

function writeSelectedLiveTargetKey(targetKey) {
  try {
    if (targetKey) {
      window.localStorage.setItem(selectedLiveTargetStorageKey, targetKey);
      return;
    }
    window.localStorage.removeItem(selectedLiveTargetStorageKey);
  } catch {
    // Local selection persistence is optional; the in-memory state remains authoritative.
  }
}

export function resolveStoredLiveTargetKey(liveTargets) {
  const storedTargetKey = readSelectedLiveTargetKey();
  return liveTargets.find((target) => target.key === storedTargetKey)?.key ?? liveTargets[0]?.key ?? "";
}

export function useLiveTargetSelection(liveTargets) {
  const [selectedLiveTargetKey, setSelectedLiveTargetKey] = useState(readSelectedLiveTargetKey);
  const selectedLiveTarget = useMemo(
    () => liveTargets.find((target) => target.key === selectedLiveTargetKey) ?? liveTargets[0] ?? null,
    [liveTargets, selectedLiveTargetKey]
  );

  useEffect(() => {
    if (liveTargets.length === 0) {
      setSelectedLiveTargetKey("");
      return;
    }
    if (!selectedLiveTargetKey || !liveTargets.some((target) => target.key === selectedLiveTargetKey)) {
      setSelectedLiveTargetKey(liveTargets[0].key);
    }
  }, [liveTargets, selectedLiveTargetKey]);

  useEffect(() => {
    writeSelectedLiveTargetKey(selectedLiveTargetKey);
  }, [selectedLiveTargetKey]);

  return {
    selectedLiveTarget,
    selectedLiveTargetKey,
    setSelectedLiveTargetKey
  };
}
