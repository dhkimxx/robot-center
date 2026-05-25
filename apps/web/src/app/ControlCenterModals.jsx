import Modal from "../components/Modal.jsx";
import { MissionFormFields } from "../domains/missions/MissionFormFields.jsx";
import { RecordingPlaybackModal } from "../domains/recordings/RecordingsScreen.jsx";
import { RobotConnectionInfoDetails, RobotFormFields } from "../domains/robots/RobotFormFields.jsx";

export function ControlCenterModals({ controller }) {
  return (
    <>
      {controller.robotModal === "create" ? (
        <Modal
          description="새 로봇을 관제 목록에 등록합니다."
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={controller.closeRobotModal}>취소</button>
              <button className="primary-button" form="robot-create-form" type="submit">등록</button>
            </>
          )}
          onClose={controller.closeRobotModal}
          title="로봇 등록"
        >
          <form className="form-grid modal-form" id="robot-create-form" onSubmit={controller.createRobot}>
            <RobotFormFields form={controller.robotForm} setForm={controller.setRobotForm} />
          </form>
        </Modal>
      ) : null}
      {controller.robotModal === "edit" && controller.selectedRobot ? (
        <Modal
          description={controller.selectedRobot.robotCode}
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={controller.closeRobotModal}>취소</button>
              <button className="primary-button" form="robot-edit-form" type="submit">저장</button>
            </>
          )}
          onClose={controller.closeRobotModal}
          title="로봇 수정"
        >
          <form className="form-grid modal-form" id="robot-edit-form" onSubmit={controller.updateRobot}>
            <RobotFormFields form={controller.robotEditForm} setForm={controller.setRobotEditForm} />
          </form>
        </Modal>
      ) : null}
      {controller.robotModal === "connection" && controller.connectionInfo ? (
        <Modal
          description={controller.connectionInfo.robotCode}
          footer={<button className="primary-button" type="button" onClick={controller.closeRobotModal}>확인</button>}
          onClose={controller.closeRobotModal}
          title="연결 정보"
          size="large"
        >
          <RobotConnectionInfoDetails connectionInfo={controller.connectionInfo} />
        </Modal>
      ) : null}
      {controller.pendingArchiveRobotCode ? (
        <Modal
          description={controller.pendingArchiveRobot?.robotCode ?? controller.pendingArchiveRobotCode}
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={() => controller.setPendingArchiveRobotCode("")}>취소</button>
              <button className="danger-button" type="button" onClick={() => void controller.confirmArchiveRobot()}>삭제</button>
            </>
          )}
          onClose={() => controller.setPendingArchiveRobotCode("")}
          title="로봇 삭제"
        >
          <div className="confirm-message">
            <strong>{controller.pendingArchiveRobot?.displayName ?? controller.pendingArchiveRobotCode}</strong>
            <span>로봇을 목록에서 제거합니다. 진행 중이거나 대기 중인 임무에 배정되어 있으면 삭제 요청은 실패합니다.</span>
          </div>
        </Modal>
      ) : null}
      {controller.missionModal === "create" ? (
        <Modal
          description="새 임무를 생성하고 로봇을 배정합니다."
          footer={(
            <>
              <button className="secondary-button" type="button" onClick={controller.closeMissionModal}>취소</button>
              <button className="primary-button" form="mission-create-form" type="submit">생성</button>
            </>
          )}
          onClose={controller.closeMissionModal}
          title="임무 생성"
        >
          <form className="form-grid modal-form" id="mission-create-form" onSubmit={controller.createMission}>
            <MissionFormFields form={controller.missionForm} robots={controller.robots} setForm={controller.setMissionForm} />
          </form>
        </Modal>
      ) : null}
      {controller.recordingPlaybackFile ? (
        <RecordingPlaybackModal
          file={controller.recordingPlaybackFile}
          onClose={() => controller.setRecordingPlaybackFile(null)}
        />
      ) : null}
    </>
  );
}
