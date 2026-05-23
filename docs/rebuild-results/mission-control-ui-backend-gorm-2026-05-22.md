---
title: "mission-control-ui-backend-gorm"
created: 2026-05-22
updated: 2026-05-22
type: "log"
tags: ["rebuild", "ui", "mission-control", "gorm", "webrtc", "verification"]
---
# Mission Control UI / Backend GORM - 2026-05-22

## 구현 범위

- UI에서 `Live 관제` 탭 제거
- `임무 목록 -> 관제 -> 임무 관제 화면 -> 로봇 선택` 흐름 적용
- 선택 로봇 기준 RGB, Thermal, 위치 지도, 센서 4분면 배치
- 관제 화면 안에서 녹화 MP4를 `재생` 버튼으로 모달 플레이어 실행
- PostgreSQL repository context 전달 정리
- `context.Background()` 잔여 제거
- GORM persistence model 추가
- 안전한 범위의 robot, mission, streaming status query/write를 GORM 기반으로 전환
- 서비스 계층 transaction 진입점 `Services.WithTransaction` 추가

## 수정 파일

- `apps/web/src/App.jsx`
- `apps/web/src/styles.css`
- `apps/server/internal/api/server.go`
- `apps/server/internal/service/services.go`
- `apps/server/internal/store/store.go`
- `apps/server/internal/store/postgres.go`
- `apps/server/internal/store/gorm_models.go`
- `docs/ai-web-prd.md`
- `docs/appendix/ui-plan.md`

## 검증 명령

```bash
cd apps/server && go test ./...
cd apps/web && npm run build
SKIP_WEB_BUILD=1 SKIP_ANDROID=1 ./scripts/dev-up.sh
MOCK_ROBOT_COUNT=3 ./scripts/python-mock-robots-up.sh
./scripts/dev-status.sh
```

## 실제 확인 결과

- app-server: OK
- recorder-worker: OK
- TURN URL: `turn:172.30.1.5:3478?transport=udp`
- SFU signaling URL: `ws://172.30.1.5:18080/sfu/ws`
- Python Mock Robot 3대 실행
- SFU room 3개 생성
- 각 recorder room에서 `trackCount=3`, `dataChannelCount=2`, `iceState=connected` 확인
- telemetry/sensor DB 적재 확인
- 브라우저에서 임무 목록, 관제 진입, 선택 로봇 4분면 화면, 녹화 재생 모달 확인

## 사용자 확인 URL

- UI: `http://172.30.1.5:18080`
- 로컬 UI: `http://127.0.0.1:18080`
- MinIO Console: `http://127.0.0.1:9001`

## 남은 한계

- Android Mock Robot은 현재 ADB device가 없어 자동 실행 검증하지 못했다.
- recording/storage 복합 흐름은 아직 raw SQL transaction을 유지한다.
- domain struct JSON tag 제거는 다음 단계 작업으로 남긴다.
