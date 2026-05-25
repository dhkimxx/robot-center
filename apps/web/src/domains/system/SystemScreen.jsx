import { componentLabels } from "../../config/controlCenterConfig.js";
import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import Surface from "../../components/ui/Surface.jsx";

export default function SystemScreen({ statusError, systemStatus }) {
  return (
    <section className="grid h-full min-h-0 grid-cols-[minmax(340px,0.75fr)_minmax(0,1fr)] gap-3 max-[980px]:grid-cols-1">
      <Surface className="self-start">
        <SectionHeader title="서비스 상태" meta={statusError ? "대기" : "정상"} />
        {statusError ? (
          <EmptyState>응답 대기: {statusError}</EmptyState>
        ) : (
          <ul className="grid gap-2">
            {(systemStatus?.components ?? []).map((component) => (
              <li
                className="flex min-h-11 items-center justify-between gap-3 rounded-lg border border-slate-500/20 bg-white/[0.045] px-3 py-2"
                key={component.name}
              >
                <span className="text-sm font-semibold text-slate-300">{componentLabels[component.name] ?? component.name}</span>
                <strong className="text-sm font-bold text-slate-50">{component.status}</strong>
              </li>
            ))}
          </ul>
        )}
      </Surface>

      <Surface className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden">
        <SectionHeader title="실시간 연결" meta={`${systemStatus?.sfuRooms?.length ?? 0}개`} />
        <div className="grid min-h-0 auto-rows-max content-start gap-2 overflow-auto pr-1">
          {(systemStatus?.sfuRooms ?? []).length === 0 ? (
            <EmptyState>연결된 세션이 없습니다.</EmptyState>
          ) : (
            systemStatus.sfuRooms.map((room) => (
              <div
                className="rounded-lg border border-slate-500/20 bg-white/[0.045] px-3 py-2"
                key={room.roomId}
              >
                <div>
                  <strong className="block truncate text-sm font-bold text-slate-50">{room.roomId}</strong>
                  <span className="mt-1 block text-xs font-semibold text-slate-400">
                    로봇 {room.robotCount}대 / 관제 {room.operatorCount}명 / 녹화 {room.recorderCount}개
                  </span>
                </div>
              </div>
            ))
          )}
        </div>
      </Surface>
    </section>
  );
}
