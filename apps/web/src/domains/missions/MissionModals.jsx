import Modal from "../../components/Modal.jsx";
import { MissionFormFields } from "./MissionFormFields.jsx";

export function MissionModals({
  createMission,
  missionForm,
  missionModal,
  onClose,
  robots,
  setMissionForm
}) {
  return missionModal === "create" ? (
    <Modal
      description="새 임무를 생성하고 로봇을 배정합니다."
      footer={(
        <>
          <button className="secondary-button" type="button" onClick={onClose}>취소</button>
          <button className="primary-button" form="mission-create-form" type="submit">생성</button>
        </>
      )}
      onClose={onClose}
      title="임무 생성"
    >
      <form className="form-grid modal-form" id="mission-create-form" onSubmit={createMission}>
        <MissionFormFields form={missionForm} robots={robots} setForm={setMissionForm} />
      </form>
    </Modal>
  ) : null;
}
