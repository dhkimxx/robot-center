export function getRecordingObjectEntries(recording) {
  if (!recording) {
    return [];
  }
  if (Array.isArray(recording.files) && recording.files.length > 0) {
    return recording.files.filter((entry) => entry.objectKey || entry.url);
  }
  const mediaObjectKeys = recording.mediaObjectKeys ?? {};
  return [
    { type: "rgb_audio_mp4", label: "RGB MP4", objectKey: mediaObjectKeys.rgbMp4, status: "planned" },
    { type: "thermal_mp4", label: "Thermal MP4", objectKey: mediaObjectKeys.thermal, status: "planned" },
    { type: "sensor_jsonl", label: "센서 기록", objectKey: mediaObjectKeys.sensor, status: "planned" },
    { type: "telemetry_jsonl", label: "위치 기록", objectKey: mediaObjectKeys.telemetry, status: "planned" },
    { type: "manifest", label: "저장 메타데이터", objectKey: recording.manifestObjectKey ?? mediaObjectKeys.manifest, status: recording.status === "uploaded" ? "available" : "recording" }
  ].filter((entry) => entry.objectKey);
}

export function makeRecordingSessionGroups(recordings) {
  const groupsByKey = new Map();
  recordings.forEach((recording) => {
    const groupKey = recording.recordingSessionId || `${recording.missionCode}-${recording.robotCode}`;
    if (!groupsByKey.has(groupKey)) {
      groupsByKey.set(groupKey, {
        id: groupKey,
        missionCode: recording.missionCode,
        robotCode: recording.robotCode,
        chunks: []
      });
    }
    groupsByKey.get(groupKey).chunks.push(recording);
  });

  return Array.from(groupsByKey.values()).map((group) => {
    const orderedChunks = group.chunks.sort((a, b) => Number(a.chunkIndex) - Number(b.chunkIndex));
    const chunks = [...orderedChunks].reverse();
    const fileEntries = orderedChunks.flatMap(getRecordingObjectEntries);
    const availableFileCount = fileEntries.filter((entry) => entry.status === "available" || entry.stored).length;
    return {
      ...group,
      chunks,
      startedAt: orderedChunks[0]?.startedAt,
      endedAt: orderedChunks[orderedChunks.length - 1]?.endedAt,
      status: orderedChunks[orderedChunks.length - 1]?.status ?? "recording",
      fileCount: fileEntries.length,
      availableFileCount
    };
  }).sort((a, b) => new Date(b.endedAt ?? b.startedAt ?? 0).getTime() - new Date(a.endedAt ?? a.startedAt ?? 0).getTime());
}

export function makeRecordingRobotGroups(sessionGroups) {
  const groupsByRobot = new Map();
  sessionGroups.forEach((session) => {
    if (!groupsByRobot.has(session.robotCode)) {
      groupsByRobot.set(session.robotCode, {
        robotCode: session.robotCode,
        missionCodes: new Set(),
        sessions: []
      });
    }
    const group = groupsByRobot.get(session.robotCode);
    group.missionCodes.add(session.missionCode);
    group.sessions.push(session);
  });

  return Array.from(groupsByRobot.values()).map((group) => {
    const sessions = group.sessions.sort((a, b) => new Date(b.endedAt ?? b.startedAt ?? 0).getTime() - new Date(a.endedAt ?? a.startedAt ?? 0).getTime());
    const latestSession = sessions[0] ?? null;
    return {
      robotCode: group.robotCode,
      sessionCount: sessions.length,
      chunkCount: sessions.reduce((total, session) => total + session.chunks.length, 0),
      latestAt: latestSession?.endedAt ?? latestSession?.startedAt,
      missionCount: group.missionCodes.size,
      sessions
    };
  }).sort((a, b) => new Date(b.latestAt ?? 0).getTime() - new Date(a.latestAt ?? 0).getTime());
}

export function makeFileStatusLabel(status) {
  const labels = {
    available: "저장됨",
    recording: "녹화 중",
    finalizing: "저장 마무리",
    planned: "예정",
    partial: "부분 저장",
    stopped: "부분 저장",
    failed: "실패"
  };
  return labels[status] ?? "대기";
}

export function makeRecordingFileAvailabilityNote(entry) {
  if (isPlayableRecordingFile(entry)) {
    return "재생 가능";
  }
  if (entry?.status === "available" && entry?.url) {
    return "파일 열기 가능";
  }
  const labels = {
    available: "파일 URL 대기",
    failed: "저장 실패",
    finalizing: "파일 업로드 대기",
    partial: "이 파일은 저장되지 않음",
    planned: "아직 생성되지 않음",
    recording: "청크 작성 중",
    stopped: "이 파일은 저장되지 않음"
  };
  return labels[entry?.status] ?? "파일 대기";
}

export function isPlayableRecordingFile(entry) {
  return Boolean(entry?.url) && entry?.status === "available" && String(entry?.contentType ?? "").startsWith("video/");
}

export function getPlayableRecordingVideoEntries(recording) {
  return getRecordingObjectEntries(recording).filter(isPlayableRecordingFile);
}

export function createRecordingPlaybackFile(recording, entry) {
  if (!recording) {
    return null;
  }
  const playableEntry = entry ?? getPlayableRecordingVideoEntries(recording)[0];
  if (!isPlayableRecordingFile(playableEntry)) {
    return null;
  }
  return {
    ...playableEntry,
    chunkIndex: recording.chunkIndex,
    endedAt: recording.endedAt,
    missionCode: recording.missionCode,
    robotCode: recording.robotCode,
    startedAt: recording.startedAt
  };
}

export function findLatestRecordingForTarget(recordings, missionCode, robotCode, predicate = () => true) {
  const hasTarget = Boolean(missionCode || robotCode);
  const matchingRecording = recordings.find((recording) => {
    if (!predicate(recording)) {
      return false;
    }
    if (missionCode && recording.missionCode !== missionCode) {
      return false;
    }
    if (robotCode && recording.robotCode !== robotCode) {
      return false;
    }
    return true;
  });

  if (matchingRecording || hasTarget) {
    return matchingRecording ?? null;
  }
  return recordings.find(predicate) ?? null;
}

export function makeRecordingStateForTarget(recordings, missionCode, robotCode) {
  const recording = findLatestRecordingForTarget(recordings, missionCode, robotCode);
  const status = recording?.status ?? "";
  if (["recording", "pending", "uploading"].includes(status)) {
    return {
      isActive: true,
      label: "녹화 중",
      recording,
      tone: "recording"
    };
  }
  if (status === "finalizing") {
    return {
      isActive: false,
      label: "저장 마무리",
      recording,
      tone: "recording"
    };
  }
  if (status === "uploaded") {
    return {
      isActive: false,
      label: "저장 완료",
      recording,
      tone: "available"
    };
  }
  if (["failed", "error"].includes(status)) {
    return {
      isActive: false,
      label: "녹화 오류",
      recording,
      tone: "danger"
    };
  }
  if (status === "partial" || status === "stopped") {
    return {
      isActive: false,
      label: "부분 저장",
      recording,
      tone: "idle"
    };
  }
  return {
    isActive: false,
    label: "녹화 대기",
    recording,
    tone: "idle"
  };
}
