import Modal from "./components/Modal.jsx";
import NotificationStack from "./components/NotificationStack.jsx";
import { navigationItems } from "./config/controlCenterConfig.js";
import { MissionFormFields } from "./domains/missions/MissionFormFields.jsx";
import MissionsScreen from "./domains/missions/MissionsScreen.jsx";
import RecordingsScreen, { RecordingPlaybackModal } from "./domains/recordings/RecordingsScreen.jsx";
import { RobotConnectionInfoDetails, RobotFormFields } from "./domains/robots/RobotFormFields.jsx";
import RobotsScreen from "./domains/robots/RobotsScreen.jsx";
import SystemScreen from "./domains/system/SystemScreen.jsx";
import { useControlCenterController } from "./hooks/useControlCenterController.js";

export default function App() {
  const {
    activeTab,
    navigateToTab,
    systemStatus,
    robots,
    missions,
    recordings,
    connectionInfo,
    statusError,
    notifications,
    robotForm,
    setRobotForm,
    selectedRobot,
    robotEditForm,
    setRobotEditForm,
    missionForm,
    setMissionForm,
    recordingPlaybackFile,
    setRecordingPlaybackFile,
    missionControlMission,
    missionControlTargets,
    selectedLiveSession,
    liveSessions,
    latestRecording,
    latestPlayableRecording,
    latestTelemetry,
    latestSensor,
    operationStatuses,
    selectedMission,
    selectedLiveTargetKey,
    setSelectedLiveTargetKey,
    robotModal,
    missionModal,
    pendingArchiveRobotCode,
    pendingArchiveRobot,
    setPendingArchiveRobotCode,
    archiveRobot,
    closeMissionModal,
    closeRobotModal,
    confirmArchiveRobot,
    connectLive,
    createMission,
    createRobot,
    disconnectLive,
    dismissNotification,
    endMission,
    loadConnectionInfo,
    openMissionControl,
    openMissionCreateModal,
    openRobotCreateModal,
    openRobotEditModal,
    playLatestRecording,
    rotateRobotToken,
    setMissionControlCode,
    setSelectedMissionManagementCode,
    setSelectedRobotManagementCode,
    startMission,
    updateRobot
  } = useControlCenterController();

  return (
    <main className="app-shell">
      <aside className="sidebar">
        <nav className="nav-list">
          {navigationItems.map((item) => (
            <button
              className={activeTab === item.key ? "nav-item active" : "nav-item"}
              key={item.key}
              type="button"
              onClick={() => navigateToTab(item.key)}
            >
              {item.label}
            </button>
          ))}
        </nav>
      </aside>

      <section className="workspace">
        <div className="workspace-header">
          <header className="topbar">
            <div>
              <p className="eyebrow">SST Robot Control PoC</p>
              <h1>Sapphire Command Center</h1>
            </div>
          </header>
          {statusError ? (
            <div className="server-alert" role="status">
              서버 응답 대기: {statusError}
            </div>
          ) : null}
        </div>

        <div className="workspace-content">
          {activeTab === "robots" ? (
            <RobotsScreen
              missions={missions}
              onArchiveRobot={archiveRobot}
              onOpenCreateRobotModal={openRobotCreateModal}
              onOpenEditRobotModal={openRobotEditModal}
              onLoadConnectionInfo={loadConnectionInfo}
              onRotateRobotToken={rotateRobotToken}
              onSelectRobot={setSelectedRobotManagementCode}
              robots={robots}
              selectedRobot={selectedRobot}
            />
          ) : null}

          {activeTab === "missions" ? (
            <MissionsScreen
              controlMission={missionControlMission}
              latestRecording={latestRecording}
              latestSensor={latestSensor}
              latestTelemetry={latestTelemetry}
              liveEvents={selectedLiveSession.events}
              liveSessions={liveSessions}
              missionTargets={missionControlTargets}
              missions={missions}
              onBackToMissionList={() => setMissionControlCode("")}
              onConnectSelectedMissionTarget={() => void connectLive()}
              onDisconnectSelectedMissionTarget={disconnectLive}
              onEndMission={endMission}
              onOpenCreateMissionModal={openMissionCreateModal}
              onOpenMissionControl={openMissionControl}
              onOpenRecordings={() => navigateToTab("recordings")}
              onPlayLatestRecording={playLatestRecording}
              onSelectMission={setSelectedMissionManagementCode}
              onStartMission={startMission}
              operationStatuses={operationStatuses}
              playbackRecording={latestPlayableRecording}
              robots={robots}
              selectedMission={selectedMission}
              selectedMissionTargetKey={selectedLiveTargetKey}
              setSelectedMissionTargetKey={setSelectedLiveTargetKey}
            />
          ) : null}

          {activeTab === "recordings" ? (
            <RecordingsScreen
              onOpenPlaybackFile={setRecordingPlaybackFile}
              recordings={recordings}
            />
          ) : null}

          {activeTab === "system" ? (
            <SystemScreen statusError={statusError} systemStatus={systemStatus} />
          ) : null}
        </div>
      </section>
      <NotificationStack notifications={notifications} onDismiss={dismissNotification} />
      {robotModal === "create" ? (
        <Modal
          description="새 로봇을 관제 목록에 등록합니다."
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={closeRobotModal}>취소</button>
              <button className="primary-button" form="robot-create-form" type="submit">등록</button>
            </>
          )}
          onClose={closeRobotModal}
          title="로봇 등록"
        >
          <form className="form-grid modal-form" id="robot-create-form" onSubmit={createRobot}>
            <RobotFormFields form={robotForm} setForm={setRobotForm} />
          </form>
        </Modal>
      ) : null}
      {robotModal === "edit" && selectedRobot ? (
        <Modal
          description={selectedRobot.robotCode}
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={closeRobotModal}>취소</button>
              <button className="primary-button" form="robot-edit-form" type="submit">저장</button>
            </>
          )}
          onClose={closeRobotModal}
          title="로봇 수정"
        >
          <form className="form-grid modal-form" id="robot-edit-form" onSubmit={updateRobot}>
            <RobotFormFields form={robotEditForm} setForm={setRobotEditForm} />
          </form>
        </Modal>
      ) : null}
      {robotModal === "connection" && connectionInfo ? (
        <Modal
          description={connectionInfo.robotCode}
          footer={<button className="primary-button" type="button" onClick={closeRobotModal}>확인</button>}
          onClose={closeRobotModal}
          title="연결 정보"
          size="large"
        >
          <RobotConnectionInfoDetails connectionInfo={connectionInfo} />
        </Modal>
      ) : null}
      {pendingArchiveRobotCode ? (
        <Modal
          description={pendingArchiveRobot?.robotCode ?? pendingArchiveRobotCode}
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={() => setPendingArchiveRobotCode("")}>취소</button>
              <button className="danger-button" type="button" onClick={() => void confirmArchiveRobot()}>삭제</button>
            </>
          )}
          onClose={() => setPendingArchiveRobotCode("")}
          title="로봇 삭제"
        >
          <div className="confirm-message">
            <strong>{pendingArchiveRobot?.displayName ?? pendingArchiveRobotCode}</strong>
            <span>로봇을 목록에서 제거합니다. 진행 중이거나 대기 중인 임무에 배정되어 있으면 삭제 요청은 실패합니다.</span>
          </div>
        </Modal>
      ) : null}
      {missionModal === "create" ? (
        <Modal
          description="새 임무를 생성하고 로봇을 배정합니다."
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={closeMissionModal}>취소</button>
              <button className="primary-button" form="mission-create-form" type="submit">생성</button>
            </>
          )}
          onClose={closeMissionModal}
          title="임무 생성"
        >
          <form className="form-grid modal-form" id="mission-create-form" onSubmit={createMission}>
            <MissionFormFields form={missionForm} robots={robots} setForm={setMissionForm} />
          </form>
        </Modal>
      ) : null}
      {recordingPlaybackFile ? (
        <RecordingPlaybackModal
          file={recordingPlaybackFile}
          onClose={() => setRecordingPlaybackFile(null)}
        />
      ) : null}
    </main>
  );
}
