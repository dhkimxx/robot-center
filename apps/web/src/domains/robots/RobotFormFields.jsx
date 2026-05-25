import { RiRefreshLine } from "react-icons/ri";

export function RobotFormFields({ form, setForm }) {
  return (
    <>
      <label className="grid gap-1.5 text-xs font-extrabold text-slate-400">
        <span>표시 이름</span>
        <input
          value={form.displayName}
          onChange={(event) => setForm({ ...form, displayName: event.target.value })}
        />
      </label>
      <label className="grid gap-1.5 text-xs font-extrabold text-slate-400">
        <span>모델</span>
        <input
          value={form.modelName}
          onChange={(event) => setForm({ ...form, modelName: event.target.value })}
        />
      </label>
    </>
  );
}

export function RobotConnectionInfoDetails({ connectionInfo, onRequestTokenReset }) {
  const rows = [
    { label: "관제 주소", value: connectionInfo.serverUrl },
    { label: "로봇 코드", value: connectionInfo.robotCode },
    {
      action: onRequestTokenReset ? (
        <button
          aria-label="로봇 토큰 초기화"
          className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-transparent text-slate-400 transition hover:border-slate-500/25 hover:bg-white/[0.06] hover:text-slate-50"
          title="토큰 초기화"
          type="button"
          onClick={() => onRequestTokenReset(connectionInfo.robotCode)}
        >
          <RiRefreshLine aria-hidden="true" />
        </button>
      ) : null,
      label: "인증 토큰",
      secret: true,
      value: connectionInfo.robotToken
    }
  ];

  return (
    <div className="grid gap-2.5">
      {rows.map((row) => (
        <div className="grid min-w-0 gap-2 rounded-xl border border-slate-500/20 bg-white/[0.045] px-3.5 py-3" key={row.label}>
          <div className="flex min-w-0 items-center justify-between gap-2.5">
            <span className="text-xs font-bold text-slate-400">{row.label}</span>
          </div>
          <div className="min-w-0">
            <code className="flex items-center justify-between gap-2 overflow-auto whitespace-nowrap rounded-lg border border-slate-500/20 bg-command-900 px-2.5 py-2 text-xs leading-relaxed text-slate-100">
              <span className="min-w-0 overflow-auto">{row.value || "-"}</span>
              {row.action}
            </code>
          </div>
        </div>
      ))}
    </div>
  );
}
