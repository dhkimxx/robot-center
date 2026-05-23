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

export function RobotConnectionInfoDetails({ connectionInfo }) {
  const rows = [
    { label: "관제 주소", value: connectionInfo.serverUrl },
    { label: "로봇 코드", value: connectionInfo.robotCode },
    { label: "인증 토큰", value: connectionInfo.robotToken, secret: true }
  ];

  return (
    <div className="connection-info-grid">
      {rows.map((row) => (
        <div className={row.secret ? "connection-info-row secret" : "connection-info-row"} key={row.label}>
          <span>{row.label}</span>
          <code>{row.value || "-"}</code>
        </div>
      ))}
    </div>
  );
}
