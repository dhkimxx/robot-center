import EmptyState from "../../components/ui/EmptyState.jsx";
import SectionHeader from "../../components/ui/SectionHeader.jsx";
import StatusBadge from "../../components/ui/StatusBadge.jsx";
import Surface from "../../components/ui/Surface.jsx";
import { ListSkeleton } from "../../components/ui/Skeleton.jsx";
import { cn } from "../../utils/cn.js";
import {
  countRoomPublishedTracks,
  countRoomRobotPublishers,
  createRoomPeerSummaries,
  makePeerLabel,
  makeRoomStreamingState
} from "./systemViewModel.js";

export default function SystemRealtimeConnections({ className = "", isInitialLoading, rooms }) {
  return (
    <Surface className={cn("grid gap-3", className)}>
      <SectionHeader title="실시간 연결" meta={isInitialLoading ? "확인 중" : `${rooms.length}개`} />
      <div className="grid auto-rows-max content-start gap-2">
        {isInitialLoading ? (
          <ListSkeleton count={4} />
        ) : rooms.length === 0 ? (
          <EmptyState>연결된 세션이 없습니다.</EmptyState>
        ) : (
          rooms.map((room) => <RealtimeConnectionCard key={room.roomId} room={room} />)
        )}
      </div>
    </Surface>
  );
}

function RealtimeConnectionCard({ room }) {
  const peerSummaries = createRoomPeerSummaries(room);
  const robotCount = countRoomRobotPublishers(room);
  const mediaCount = countRoomPublishedTracks(room);
  const roomState = makeRoomStreamingState(room);

  return (
    <div className="grid gap-3 rounded-xl border border-slate-500/20 bg-white/[0.045] p-4">
      <div className="flex min-w-0 items-start justify-between gap-3 max-[760px]:grid">
        <div className="min-w-0">
          <strong className="block truncate text-base font-black text-slate-50">{room.roomId}</strong>
          <span className="mt-1 block text-xs font-semibold text-slate-500">실시간 관제 연결</span>
        </div>
        <div className="flex flex-wrap justify-end gap-1.5 max-[760px]:justify-start">
          <StatusBadge tone={robotCount > 0 ? "success" : "neutral"}>로봇 {robotCount}대</StatusBadge>
          <StatusBadge tone={room.operatorCount > 0 ? "info" : "neutral"}>관제 {room.operatorCount}명</StatusBadge>
          <StatusBadge tone={room.recorderCount > 0 ? "warning" : "neutral"}>녹화 {room.recorderCount}개</StatusBadge>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-2 max-[760px]:grid-cols-1">
        <RoomMetric label="미디어" value={`${mediaCount}개`} />
        <RoomMetric label="연결 주체" value={`${peerSummaries.length}개`} />
        <RoomMetric label="상태" value={roomState} />
      </div>

      <div className="flex flex-wrap gap-1.5">
        {peerSummaries.map((peer) => (
          <span
            className={cn(
              "inline-flex min-h-7 max-w-full items-center rounded-full border px-2.5 text-xs font-semibold",
              peer.role === "robot" && "border-emerald-400/25 bg-emerald-400/[0.10] text-emerald-100",
              peer.role === "operator" && "border-sapphire-400/25 bg-sapphire-400/[0.10] text-sapphire-100",
              peer.role === "recorder" && "border-amber-300/25 bg-amber-300/[0.10] text-amber-100"
            )}
            key={peer.peerId}
          >
            <span className="truncate">{makePeerLabel(peer)}</span>
          </span>
        ))}
      </div>
    </div>
  );
}

function RoomMetric({ label, value }) {
  return (
    <div className="rounded-lg border border-slate-500/20 bg-command-900/50 px-3 py-2">
      <span className="block text-xs font-bold text-slate-500">{label}</span>
      <strong className="mt-1 block text-sm font-black text-slate-100">{value}</strong>
    </div>
  );
}
