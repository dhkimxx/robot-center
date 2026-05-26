export default function PageContextBar({ action, meta, title }) {
  return (
    <section className="flex min-h-[52px] min-w-0 items-center justify-between gap-4 border-b border-slate-800/70 bg-command-900/40 px-4">
      <div className="min-w-0">
        <h2 className="truncate text-sm font-bold text-slate-100">{title}</h2>
        {meta ? <p className="mt-0.5 truncate text-xs font-semibold text-slate-500">{meta}</p> : null}
      </div>
      {action ? <div className="flex shrink-0 items-center gap-2">{action}</div> : null}
    </section>
  );
}
