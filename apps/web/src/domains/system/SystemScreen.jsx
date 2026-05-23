import { componentLabels } from "../../config/controlCenterConfig.js";

export default function SystemScreen({ statusError, systemStatus }) {
  return (
    <section className="two-column-grid">
      <article className="surface">
        <div className="section-heading">
          <h2>서비스 상태</h2>
          <span>{statusError ? "대기" : "정상"}</span>
        </div>
        {statusError ? (
          <p className="empty-state">응답 대기: {statusError}</p>
        ) : (
          <ul className="component-list">
            {(systemStatus?.components ?? []).map((component) => (
              <li key={component.name}>
                <span>{componentLabels[component.name] ?? component.name}</span>
                <strong>{component.status}</strong>
              </li>
            ))}
          </ul>
        )}
      </article>

      <article className="surface">
        <div className="section-heading">
          <h2>실시간 연결</h2>
          <span>{systemStatus?.sfuRooms?.length ?? 0}개</span>
        </div>
        <div className="list-block">
          {(systemStatus?.sfuRooms ?? []).length === 0 ? (
            <p className="empty-state">연결된 세션이 없습니다.</p>
          ) : (
            systemStatus.sfuRooms.map((room) => (
              <div className="row-item" key={room.roomId}>
                <div>
                  <strong>{room.roomId}</strong>
                  <span>로봇 {room.robotCount}대 / 관제 {room.operatorCount}명 / 녹화 {room.recorderCount}개</span>
                </div>
              </div>
            ))
          )}
        </div>
      </article>
    </section>
  );
}
