# Sapphire Command Center UI Renewal Result

## 구현 범위

- 관제 UI 첫 진입 화면을 `진행 임무` 중심으로 변경했다.
- SST PoC 시연 톤에 맞춰 딥 네이비/블랙 기반, 사파이어 블루 포인트, 굵은 타이포 중심의 `Sapphire Command Center` 스타일을 적용했다.
- Tailwind/PostCSS 설정과 `clsx`/`tailwind-merge` 기반 유틸을 추가했다.
- `App.jsx`의 API 호출, formatter, mission/live/recording/robot helper를 도메인 파일로 분리했다.
- 임무 관제 화면을 mission 중심으로 유지하고, 선택 로봇의 `RGB / Thermal / 위치 / 센서` 4분면이 주 시각 요소가 되도록 조정했다.
- 녹화 화면을 로봇별 세션 목록과 청크 카드 구조로 정리하고, MP4는 `재생`, JSON/manifest는 `열기/보기` 액션으로 분리했다.
- 렌더 예외 시 화면 전체가 비지 않도록 앱 레벨 에러 바운더리를 추가했다.
- Vite 개발 프록시를 현재 시연 기준 포트 `http://127.0.0.1:18080`으로 맞췄다.

## 수정 파일

- `apps/web/package.json`
- `apps/web/package-lock.json`
- `apps/web/postcss.config.js`
- `apps/web/tailwind.config.js`
- `apps/web/vite.config.js`
- `apps/web/src/main.jsx`
- `apps/web/src/App.jsx`
- `apps/web/src/styles.css`
- `apps/web/src/styles/base.css`
- `apps/web/src/styles/tokens.css`
- `apps/web/src/lib/cn.js`
- `apps/web/src/api/controlCenterApi.js`
- `apps/web/src/domain/config.js`
- `apps/web/src/domain/formatters.js`
- `apps/web/src/domain/live/liveHelpers.js`
- `apps/web/src/domain/missions/missionHelpers.js`
- `apps/web/src/domain/recordings/recordingHelpers.js`
- `apps/web/src/domain/robots/robotHelpers.js`

## 검증 명령

```bash
cd apps/web && npm run build
./scripts/dev-status.sh
```

## 실제 확인 결과

- `npm run build` 통과.
- `http://127.0.0.1:18080` 접속 시 첫 화면이 `진행 임무`로 표시됨.
- `mission-007` 관제 진입 확인.
- `전체 연결` 후 `robot-001`, `robot-002` 모두 연결됨으로 표시됨.
- RGB/Thermal 영상 수신 확인.
- 위치 지도와 GPS 좌표 표시 확인.
- 센서값 표시 확인.
- 녹화 화면에서 로봇별 세션/청크 목록 표시 확인.
- 녹화 MP4 `재생` 모달 표시 확인.

## 현재 시연 상태

- 사용자 확인 URL: `http://127.0.0.1:18080`
- 외부 단말 확인 URL: `http://172.30.1.5:18080`
- app-server: OK
- recorder-worker: OK
- TURN: running
- PostgreSQL/PostGIS: healthy
- MinIO: healthy
- Python mock robot: `robot-001`, `robot-002` running
- Android Mock Robot: ADB unavailable

## 남은 한계

- UI 컴포넌트 분리는 1차 helper/API/domain 분리까지 완료했다. `MissionControlPage`, `RecordingsView`, `RobotsView`를 별도 파일로 분리하는 작업은 다음 단계로 남아 있다.
- 기존 `styles.css`는 token/base와 override 중심으로 정리했지만, 완전한 Tailwind-first 컴포넌트 이전은 다음 단계에서 이어가야 한다.
- Android Mock Robot 실기기 검증은 현재 ADB unavailable 상태라 수행하지 못했다.
