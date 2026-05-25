export function MetricTile({ label, value, unit }) {
  return (
    <div className="grid content-between rounded-lg border border-slate-500/20 bg-white/[0.045] p-3">
      <span className="text-xs font-bold text-slate-400">{label}</span>
      <strong className="mt-2 text-2xl font-extrabold leading-tight text-slate-50">{value}</strong>
      <small className="mt-1 text-xs font-semibold text-slate-500">{unit}</small>
    </div>
  );
}
