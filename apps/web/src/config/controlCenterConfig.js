export const navigationItems = [
  { key: "missions", label: "진행 임무", path: "/missions" },
  { key: "robots", label: "로봇", path: "/robots" },
  { key: "system", label: "시스템", path: "/system" }
];

export const missionTypes = [
  { value: "mountain_rescue", label: "산악조난" },
  { value: "collapse_site", label: "붕괴현장" },
  { value: "underground_facility", label: "지하시설" }
];

export const componentLabels = {
  "app-server": "관제 서비스",
  "recorder-worker": "녹화 서비스",
  turn: "통신 릴레이",
  postgres: "운영 데이터",
  minio: "영상 저장"
};
