export function canApplyLiveAttempt(session, attemptId) {
  return !attemptId || !session?.attemptId || session.attemptId === attemptId;
}

export function applyLiveAttemptUpdate(session, attemptId, updater, options = {}) {
  if (!options.replaceAttempt && !canApplyLiveAttempt(session, attemptId)) {
    return session;
  }
  const nextSession = updater(session);
  if (!attemptId || nextSession === session) {
    return nextSession;
  }
  return {
    ...nextSession,
    attemptId
  };
}
