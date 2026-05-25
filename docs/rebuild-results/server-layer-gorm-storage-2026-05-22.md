---
title: "server-layer-gorm-storage"
created: 2026-05-22
updated: '2026-05-22'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>", "dhkimxx <dhkimxx@naver.com>"]
type: "log"
tags: ["rebuild", "server", "gorm", "recording", "storage-objects", "test-agent"]
history:
- "2026-05-22 danya.kim <danya.kim@thundersoft.com>: initial entry"
- '2026-05-22 dhkimxx <dhkimxx@naver.com>: added final web build and Playwright verification results'
---
# Server Layer / GORM / Storage Objects - 2026-05-22

## 범위

harness 문서와 구현 사이의 차이를 줄이기 위해 서버 구조를 정리했다.

구현 범위:

- `store.Store`를 역할별 repository interface로 분리
- `Controller -> Service -> Repository` 경계 추가
- API response DTO 추가
- PostgreSQL 연결을 GORM 기반으로 전환
- 기존 PostgreSQL/PostGIS raw SQL은 `*gorm.DB.DB()`에서 얻은 `*sql.DB`로 단계 유지
- 녹화 파일 업로드 완료 시 `storage_objects`에 object metadata upsert
- manifest 업로드 완료 시 `recording_chunks.manifest_object_id` 연결
- recorder-worker가 업로드 완료 알림에 `sizeBytes` 전달
- harness 문서에 relay-only, staged GORM, `storage_objects` 기준 반영
- Test / Regression 에이전트 관점 추가

## 수정 파일

- `apps/server/go.mod`
- `apps/server/go.sum`
- `apps/server/internal/api/server.go`
- `apps/server/internal/api/dto/dto.go`
- `apps/server/internal/service/services.go`
- `apps/server/internal/store/store.go`
- `apps/server/internal/store/memory.go`
- `apps/server/internal/store/postgres.go`
- `apps/server/internal/recording/worker.go`
- `apps/server/internal/recording/media_writer.go`
- `docs/harness/20260522-server-architecture.md`
- `docs/harness/20260522-webrtc-sfu-topology.md`
- `docs/harness/20260522-go-gorm-persistence.md`
- `docs/harness/20260522-harness-index.md`

## 검증 명령

```bash
cd apps/server
go test ./...
```

결과:

```text
ok robot-center/apps/server/internal/api
ok robot-center/apps/server/internal/recording
ok robot-center/apps/server/internal/sfu
```

```bash
SKIP_WEB_BUILD=1 ./scripts/dev-up.sh
MOCK_ROBOT_COUNT=3 ./scripts/python-mock-robots-up.sh
./scripts/dev-status.sh
```

결과:

```text
app-server ok
recorder-worker ok
system ok
3 python mock robot sessions running
3 SFU rooms connected
each recorder room: trackCount=3, dataChannelCount=2
```

```bash
cd apps/web
npm run build
```

결과:

```text
vite build successful
```

Playwright 확인:

```text
http://127.0.0.1:18080
Page Title: AI Web Control Center
로봇 관제 통합 시연 / 연결됨 / 진행 임무 3건 / 녹화 청크 240개 표시
```

## DB / MinIO 확인

```sql
select object_type, count(*) as rows, count(size_bytes) as rows_with_size, max(size_bytes) as max_size
from storage_objects
group by object_type
order by object_type;
```

결과:

```text
manifest        rows=4 rows_with_size=3
rgb_audio_mp4   rows=4 rows_with_size=3
sensor_jsonl    rows=4 rows_with_size=3
telemetry_jsonl rows=4 rows_with_size=3
thermal_mp4     rows=4 rows_with_size=3
```

신규 업로드된 row는 `bucket`, `object_key`, `content_type`, `size_bytes`, `recording_chunk_id`가 기록됐다.

중복 object key 확인:

```text
duplicate_object_keys = 0
```

MP4 URL 확인:

```text
HTTP/1.1 200 OK
Content-Type: video/mp4
```

## 사용자 확인 상태

현재 실행 상태:

- 관제 UI: `http://192.168.20.11:18080`
- 로컬 UI: `http://127.0.0.1:18080`
- MinIO console: `http://127.0.0.1:9001`
- app-server / recorder-worker / TURN 실행 중
- Python Mock Robot 3대 실행 중

## 남은 한계

- Android Mock Robot은 이번 검증에서 ADB device unavailable 상태라 실행 확인하지 못했다.
- 기존에 이미 저장된 오래된 chunk는 `storage_objects.size_bytes`와 `manifest_object_id`가 없을 수 있다.
- repository interface는 분리했지만 Postgres 구현 파일은 아직 하나다. 다음 단계에서 domain별 repository 파일로 나누는 게 좋다.
- Service layer는 현재 facade 단계다. 여러 repository를 묶는 transaction은 후속 단계에서 Service 소유로 옮긴다.
