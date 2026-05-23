---
title: "robot-crud"
created: 2026-05-22
updated: 2026-05-22
type: "log"
tags: ["robot", "crud", "ui", "api", "postgres"]
---
# Robot CRUD - 2026-05-22

## 구현 범위

- 로봇 수정 API 추가: `PATCH /api/robots/{robotCode}`
- 로봇 삭제 API 추가: `DELETE /api/robots/{robotCode}`
- 로봇 연결 토큰 재발급 API 추가: `POST /api/robots/{robotCode}/connection-token`
- PostgreSQL `robots.archived_at` soft delete 컬럼 추가
- 로봇 목록 조회 시 soft deleted robot 제외
- active/ready mission에 배정된 robot 삭제 차단
- 로봇 화면에 목록, 선택, 상세 수정, 연결 정보 조회, 토큰 재발급, 삭제 액션 추가
- `dev-up.sh`가 demo robot 기준 active/ready mission만 선택하도록 수정

## 검증

```bash
cd apps/server && go test ./...
cd apps/web && npm run build
bash -n scripts/dev-up.sh scripts/python-mock-robots-up.sh
```

Runtime smoke:

```text
POST /api/robots
PATCH /api/robots/{robotCode}
POST /api/robots/{robotCode}/connection-token
DELETE /api/robots/{robotCode}
GET /api/robots
```

결과:

- Go test 통과
- Web build 통과
- Runtime robot CRUD smoke 통과
- UI 로봇 화면에서 상세 수정/토큰 재발급/삭제 액션 확인
- app-server, recorder-worker, TURN, Python Mock Robot 3대 실행 중

## 한계

- 삭제는 물리 삭제가 아니라 `archived_at` 기반 목록 제외다.
- 진행 중 또는 대기 중 임무에 배정된 로봇은 삭제할 수 없다.
