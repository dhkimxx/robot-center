import { useState } from "react";
import ConfirmDialog from "../../components/ConfirmDialog.jsx";
import Modal from "../../components/Modal.jsx";
import Button from "../../components/ui/Button.jsx";
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
  const [pendingTokenResetRobotCode, setPendingTokenResetRobotCode] = useState("");

  const closeArchiveConfirm = () => setPendingArchiveRobotCode("");
  const closeTokenResetConfirm = () => setPendingTokenResetRobotCode("");
  const confirmTokenReset = () => {
    const robotCode = pendingTokenResetRobotCode;
    if (!robotCode) {
      return;
    }
    setPendingTokenResetRobotCode("");
    void rotateRobotToken(robotCode);
  };

  return (
    <>
      {robotModal === "create" ? (
        <Modal
          description="새 로봇을 관제 목록에 등록합니다."
          footer={(
            <>
              <Button onClick={closeRobotModal}>취소</Button>
              <Button form="robot-create-form" type="submit" variant="primary">등록</Button>
            </>
          )}
          onClose={closeRobotModal}
          title="로봇 등록"
        >
          <form className="grid gap-3" id="robot-create-form" onSubmit={createRobot}>
            <RobotFormFields form={robotForm} setForm={setRobotForm} />
          </form>
        </Modal>
      ) : null}
      {robotModal === "edit" && selectedRobot ? (
        <Modal
          description={selectedRobot.robotCode}
          footer={(
            <>
              <Button onClick={closeRobotModal}>취소</Button>
              <Button form="robot-edit-form" type="submit" variant="primary">저장</Button>
            </>
          )}
          onClose={closeRobotModal}
          title="로봇 수정"
        >
          <form className="grid gap-3" id="robot-edit-form" onSubmit={updateRobot}>
            <RobotFormFields form={robotEditForm} setForm={setRobotEditForm} />
          </form>
        </Modal>
      ) : null}
      {robotModal === "connection" && connectionInfo ? (
        <Modal
          description={connectionInfo.robotCode}
          footer={<Button variant="primary" onClick={closeRobotModal}>확인</Button>}
          onClose={closeRobotModal}
          title="연결 정보"
          size="large"
        >
          <RobotConnectionInfoDetails connectionInfo={connectionInfo} onRequestTokenReset={setPendingTokenResetRobotCode} />
        </Modal>
      ) : null}
      {pendingArchiveRobotCode ? (
        <ConfirmDialog
          confirmLabel="삭제"
          description="로봇을 목록에서 제거합니다. 진행 중이거나 대기 중인 임무에 배정되어 있으면 삭제 요청은 실패합니다."
          onCancel={closeArchiveConfirm}
          onConfirm={confirmArchiveRobot}
          subject={pendingArchiveRobot?.displayName ?? pendingArchiveRobotCode}
          title="로봇 삭제"
          tone="danger"
        />
      ) : null}
      {pendingTokenResetRobotCode ? (
        <ConfirmDialog
          confirmLabel="재발급"
          description="기존 토큰은 더 이상 사용할 수 없습니다. 로봇 쪽 연결 정보도 새 토큰으로 다시 설정해야 합니다."
          onCancel={closeTokenResetConfirm}
          onConfirm={confirmTokenReset}
          subject={pendingTokenResetRobotCode}
          title="토큰 재발급"
          tone="danger"
        />
      ) : null}
    </>
  );
}
