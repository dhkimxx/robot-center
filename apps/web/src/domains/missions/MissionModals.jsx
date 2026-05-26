import Modal from "../../components/Modal.jsx";
import Button from "../../components/ui/Button.jsx";
import { MissionFormFields } from "./MissionFormFields.jsx";

export function MissionModals({
  createMission,
  missionForm,
  missionModal,
  missions,
  observedStreams,
  onClose,
  robots,
  setMissionForm
}) {
  return missionModal === "create" ? (
    <Modal
      description="새 임무를 생성하고 로봇을 배정합니다."
      footer={(
        <>
          <Button onClick={onClose}>취소</Button>
          <Button form="mission-create-form" type="submit" variant="primary">생성</Button>
        </>
      )}
      onClose={onClose}
      title="임무 생성"
    >
      <form className="grid gap-3" id="mission-create-form" onSubmit={createMission}>
        <MissionFormFields
          form={missionForm}
          missions={missions}
          observedStreams={observedStreams}
          robots={robots}
          setForm={setMissionForm}
        />
      </form>
    </Modal>
  ) : null;
}
