import Modal from "../../components/Modal.jsx";
import { RobotConnectionInfoDetails, RobotFormFields } from "./RobotFormFields.jsx";

export function RobotModals({
  closeRobotModal,
  confirmArchiveRobot,
  connectionInfo,
  createRobot,
  pendingArchiveRobot,
  pendingArchiveRobotCode,
  robotEditForm,
  robotForm,
  robotModal,
  rotateRobotToken,
  selectedRobot,
  setPendingArchiveRobotCode,
  setRobotEditForm,
  setRobotForm,
  updateRobot
}) {
  return (
    <>
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
          <RobotConnectionInfoDetails connectionInfo={connectionInfo} onRotateToken={rotateRobotToken} />
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
    </>
  );
}
