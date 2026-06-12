import Surface from "../../../components/ui/Surface.jsx";
import { cn } from "../../../utils/cn.js";

const controlStatusSlots = [
  {
    fallbackDetail: "브라우저 관제 연결 대기",
    fallbackValue: "대기",
    key: "operator",
    label: "관제 연결"
  },
  {
    fallbackDetail: "로봇 영상 송출 대기",
    fallbackValue: "대기",
    key: "media",
    label: "로봇 송출"
  },
  {
    fallbackDetail: "센서/이벤트 수신 대기",
    fallbackValue: "대기",
    key: "data",
    label: "센서 수신"
  },
  {
    fallbackDetail: "녹화 구간 대기",
    fallbackValue: "대기",
    key: "recording",
    label: "녹화 상태"
  }
];

const statusCardToneClasses = {
  danger: {
    card: "border-red-400/35 bg-red-400/[0.09]",
    dot: "bg-red-300 shadow-red-300/40",
    value: "text-red-100"
  },
  ok: {
    card: "border-emerald-400/35 bg-emerald-400/[0.09]",
    dot: "bg-emerald-300 shadow-emerald-300/40",
    value: "text-emerald-100"
  },
  waiting: {
    card: "border-amber-300/35 bg-amber-300/[0.09]",
    dot: "bg-amber-200 shadow-amber-200/40",
    value: "text-amber-100"
  }
};

export function ControlStatusSummary({ diagnostics }) {
  const summaryItems = makeControlStatusItems(diagnostics);

  return (
    <Surface
      className="grid min-h-[76px] min-w-0 grid-cols-4 gap-2 overflow-hidden px-3 py-2.5 max-[640px]:grid-cols-2"
      data-testid="live-control-status-summary"
    >
      {summaryItems.map((item) => (
        <StatusCard
          detail={item.detail}
          key={item.key}
          label={item.label}
          tone={item.tone}
          value={item.value}
        />
      ))}
    </Surface>
  );
}

function makeControlStatusItems(diagnostics = []) {
  const diagnosticsByKey = new Map(diagnostics.map((diagnostic) => [diagnostic.key, diagnostic]));
  return controlStatusSlots.map((slot) => {
    const diagnostic = diagnosticsByKey.get(slot.key);
    return {
      detail: diagnostic?.detail ?? slot.fallbackDetail,
      key: slot.key,
      label: slot.label,
      tone: normalizeStatusTone(diagnostic?.tone),
      value: diagnostic?.value ?? slot.fallbackValue
    };
  });
}

function StatusCard({ detail, label, tone, value }) {
  const toneClasses = statusCardToneClasses[tone] ?? statusCardToneClasses.waiting;

  return (
    <div
      className={cn(
        "grid min-h-[54px] min-w-0 grid-rows-[auto_auto] content-center gap-1 rounded-lg border px-2.5 py-2",
        toneClasses.card
      )}
      title={detail}
    >
      <div className="flex min-w-0 items-center gap-1.5">
        <span
          aria-hidden
          className={cn(
            "h-2 w-2 shrink-0 rounded-full shadow-[0_0_10px]",
            toneClasses.dot
          )}
        />
        <span className="truncate text-[11px] font-bold text-slate-400">{label}</span>
      </div>
      <strong className={cn("truncate text-base font-extrabold leading-tight", toneClasses.value)}>{value}</strong>
    </div>
  );
}

function normalizeStatusTone(tone) {
  if (tone === "ok" || tone === "danger") {
    return tone;
  }
  return "waiting";
}
