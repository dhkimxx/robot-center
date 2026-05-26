# SFU 선택 구독 및 센서 Descriptor/Sample 전환 결과

## 구현 범위

- app-server 내부 SFU의 operator 선택 구독 흐름 보강
  - operator URL의 `robotCode`는 identity로 쓰지 않음
  - `select-robot`이 subscriber session 생성 전에 도착해도 선택 로봇이 session에 반영됨
  - recorder는 기존처럼 mission room 전체 robot stream을 수신
- `SensorDescriptor` / `SensorSample` 도메인, GORM AutoMigrate 모델, 저장소/API 추가
  - `GET /api/sensor-descriptors`
  - `POST /api/sensor-descriptors`
  - `GET /api/sensor-samples`
  - `POST /api/sensor-samples`
  - `GET /api/sensor-latest`
- recorder-worker DataChannel 저장 정책 수정
  - `channel.telemetry` -> sensor sample 저장
  - `channel.spatial` -> sensor sample 저장
  - `channel.event`, `channel.control`은 sensor sample 저장 제외
- Python mock robot 참조 구현 보강
  - canonical track/channel 사용
  - telemetry/spatial descriptor/sample payload 전송
  - README에 로봇 연결 순서 기록
- Web UI sensor-latest 연동
  - 선택 mission + robot 기준 최신 센서 조회
  - position sensor를 지도 패널에 연결
  - sensor descriptor/sample 기반 동적 센서 패널 표시

## 데이터 초기화

검증 전 로컬 개발 데이터만 초기화했다.

- PostgreSQL: Docker volume `robot-center_postgres-data` 삭제
- MinIO: Docker volume `robot-center_minio-data` 삭제

재생성 절차:

```bash
./scripts/dev-down.sh
docker compose -f deploy/docker-compose.yml down -v
SKIP_WEB_BUILD=1 ./scripts/dev-up.sh
./scripts/python-mock-robots-up.sh
```

## 검증 명령

```bash
cd apps/server && go test ./... && go vet ./...
cd apps/web && npm test -- --run && npm run build
python3 -m py_compile apps/mock-robot-python/mock_robot.py
./scripts/dev-status.sh
```

## 실제 확인 결과

- UI: `http://192.168.20.32:18080`
- app-server: OK
- recorder-worker: OK
- PostgreSQL: healthy
- MinIO: healthy
- Python mock robot:
  - `robot-001 -> mission-002`
  - `robot-002 -> mission-002`
- SFU room:
  - `mission-002`
  - robot publisher 2대
  - recorder subscriber 1개
  - ICE connected
  - trackCount 6
  - dataChannelCount 4
- DB 적재:
  - `sensor_descriptors`: 12 rows
  - `sensor_samples`: `channel.telemetry`, `channel.spatial` 저장 확인
- Web UI:
  - mission 관제 진입 확인
  - robot-001 자동 연결 확인
  - robot-002 선택 전환 확인
  - RGB/Thermal 표시 확인
  - GPS 위치 표시 확인
  - 센서 패널 표시 확인

## 남은 한계

- `channel.control`은 예약 채널이며 제어 포워딩, 권한, ACK/NACK, audit log는 구현하지 않았다.
- `channel.event`는 sensor sample과 분리했지만 별도 event log 저장 구조는 이번 범위에서 확정하지 않았다.
- point cloud는 descriptor와 object_ref 형태만 준비했고, 실제 MinIO object 저장은 다음 단계로 남겼다.
