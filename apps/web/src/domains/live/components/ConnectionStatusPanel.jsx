export function ConnectionStatusPanel({ compact = false, statuses }) {
  return (
    <article className={compact ? "status-surface embedded" : "surface status-surface"}>
      <div className="section-heading">
        <h2>연결 상태</h2>
        <span>{statuses.filter((status) => status.tone === "ok").length}/{statuses.length}</span>
      </div>
      <div className="status-grid">
        {statuses.map((status) => (
          <div className={`status-cell ${status.tone}`} key={status.label}>
            <span>{status.label}</span>
            <strong>{status.value}</strong>
            <small>{status.detail}</small>
          </div>
        ))}
      </div>
    </article>
  );
}
