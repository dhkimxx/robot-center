export function RobotFormFields({ form, setForm }) {
  return (
    <>
      <label>
        표시 이름
        <input
          value={form.displayName}
          onChange={(event) => setForm({ ...form, displayName: event.target.value })}
        />
      </label>
      <label>
        모델
        <input
          value={form.modelName}
          onChange={(event) => setForm({ ...form, modelName: event.target.value })}
        />
      </label>
    </>
  );
}

export function RobotConnectionInfoDetails({ connectionInfo, onRotateToken }) {
  const rows = [
    { label: "관제 주소", value: connectionInfo.serverUrl },
    { label: "로봇 코드", value: connectionInfo.robotCode },
    {
      action: onRotateToken ? (
        <button
          className="secondary-button compact-button"
          type="button"
          onClick={() => void onRotateToken(connectionInfo.robotCode)}
        >
          재발급
        </button>
      ) : null,
      label: "인증 토큰",
      secret: true,
      value: connectionInfo.robotToken
    }
  ];

  return (
    <div className="connection-info-grid">
      {rows.map((row) => (
        <div className={row.secret ? "connection-info-row secret" : "connection-info-row"} key={row.label}>
          <div className="connection-info-row-header">
            <span>{row.label}</span>
            {row.action}
          </div>
          <code>{row.value || "-"}</code>
        </div>
      ))}
    </div>
  );
}
