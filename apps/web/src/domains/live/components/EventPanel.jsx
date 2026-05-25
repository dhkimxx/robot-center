import { formatDateTime } from "../../../utils/formatters.js";

export function EventPanel({ liveEvents }) {
  return (
    <article className="surface event-panel">
      <div className="section-heading">
        <h2>이벤트</h2>
        <span>{liveEvents.length}건</span>
      </div>
      <div className="event-list">
        {liveEvents.length === 0 ? (
          <p className="empty-state">관제 연결 이벤트가 없습니다.</p>
        ) : (
          liveEvents.map((event) => (
            <div className="event-row" key={event.id}>
              <span>{formatDateTime(event.at)}</span>
              <strong>{event.message}</strong>
            </div>
          ))
        )}
      </div>
    </article>
  );
}
