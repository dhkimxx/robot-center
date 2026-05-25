import { useEffect, useRef } from "react";
import { makeLiveTargetKey } from "./liveHelpers.js";
import {
  activeLiveConnectionStatuses,
  reconnectableLiveStatuses
} from "./liveConnectionStates.js";

export function useLiveAutoConnect({
  activeSection,
  connectLiveTarget,
  currentConnection,
  missionControlMission,
  selectedLiveSession,
  selectedLiveTarget
}) {
  const autoConnectingTargetKeyRef = useRef("");

  useEffect(() => {
    if (activeSection !== "missions" || !missionControlMission || missionControlMission.status !== "active" || !selectedLiveTarget) {
      return;
    }

    const targetKey = makeLiveTargetKey(selectedLiveTarget);

    if (
      reconnectableLiveStatuses.has(selectedLiveSession.status)
      && selectedLiveSession.events.length > 0
      && (!currentConnection || currentConnection.targetKey === targetKey)
    ) {
      return;
    }
    if (currentConnection?.targetKey === targetKey && activeLiveConnectionStatuses.has(selectedLiveSession.status)) {
      return;
    }
    if (autoConnectingTargetKeyRef.current === targetKey) {
      return;
    }

    autoConnectingTargetKeyRef.current = targetKey;
    void connectLiveTarget(selectedLiveTarget).finally(() => {
      if (autoConnectingTargetKeyRef.current === targetKey) {
        autoConnectingTargetKeyRef.current = "";
      }
    });
  }, [activeSection, connectLiveTarget, currentConnection, missionControlMission, selectedLiveSession, selectedLiveTarget]);
}
