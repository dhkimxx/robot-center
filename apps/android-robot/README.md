# Android Mock Robot

P0 제품형 구조에서 Android 단말을 Robot Gateway 샘플로 사용하는 앱이다.

앱은 관제센터에서 발급한 아래 값을 입력받는다.

```text
serverUrl
robotCode
robotToken
```

그 다음 REST API로 heartbeat와 mission 조회를 수행하고, active mission을 받으면 app-server/SFU로 WebRTC publish를 시작한다.
위치, 가스, 배터리 데이터는 REST로 직접 저장하지 않고 `channel.telemetry` DataChannel로만 보낸다.

Mission 단위 multi-robot 구조에서는 여러 Android Mock Robot이 같은 `missionCode` room으로 publish한다. 각 앱 인스턴스는 서로 다른 `robotCode`와 `robotToken`으로 실행하되, telemetry/spatial/event payload에는 `robotCode`, `missionId`, `missionCode`, `channelRole`을 넣지 않는다. 서버가 token, room, DataChannel label에서 context를 주입한다.

## 송신 범위

- RGB: Android 실제 카메라
- Thermal: 앱 내부 synthetic thermal video track
- Audio: Android microphone Opus track
- Telemetry: `channel.telemetry` DataChannel, Android LocationManager 기반 GPS, 1Hz
- Spatial: `channel.spatial` DataChannel, negotiation 확인용. payload schema는 아직 송신하지 않는다.
- Event: `channel.event` DataChannel, negotiation 확인용
- Control: `channel.control` DataChannel reserved, negotiation 확인용
- Video codec: H.264 우선
- ICE 정책: TURN relay only

## 관제센터 연동 흐름

```text
1. React UI에서 로봇 생성
2. UI에서 serverUrl, robotCode, robotToken 확인
3. Android Mock Robot에 값 입력
4. Connect Center
5. POST /api/v1/robot/heartbeat
6. GET /api/v1/robot/mission
7. active mission이면 mission 응답의 `sfu.signalingUrl`로 app-server/SFU WebRTC publish
8. `channel.telemetry` OPEN 이후 telemetry payload 송신
9. recorder-worker가 SFU subscriber로 telemetry를 받아 저장
```

## 빌드

```bash
cd /Users/dhkim/workspace/sst/robot-center/apps/android-robot
./gradlew --no-daemon :app:assembleDebug
```

APK 위치:

```text
/Users/dhkim/workspace/sst/robot-center/apps/android-robot/app/build/outputs/apk/debug/app-debug.apk
```

## 단말 설치

```bash
adb devices -l
adb install -r app/build/outputs/apk/debug/app-debug.apk
adb shell am start -n com.sst.robotcenter.androidrobot/.MainActivity
```

## 실제 Android 단말 IP 주의

Android 단말에서 `127.0.0.1`은 Mac이 아니라 단말 자신이다. 실제 단말에서는 Mac의 LAN IP를 사용한다.

Mac IP 확인:

```bash
ipconfig getifaddr en0
```

예시:

```text
serverUrl = http://192.168.20.26:18080
```

## Intent extra 실행

UI에서 발급받은 값을 adb로 바로 주입할 수 있다. 여러 로봇을 같은 mission에 붙일 때는 단말 또는 에뮬레이터별로 서로 다른 `robotCode`, `robotToken`을 주입한다.

```bash
adb shell 'am start -n com.sst.robotcenter.androidrobot/.MainActivity \
  --es serverUrl "http://192.168.20.26:18080" \
  --es robotCode "robot-001" \
  --es robotToken "rb_p0_xxxxx"'
```

권한이 이미 허용된 상태면 자동 연결도 가능하다.

```bash
adb shell input keyevent KEYCODE_WAKEUP
adb shell wm dismiss-keyguard
adb shell 'am start -W -n com.sst.robotcenter.androidrobot/.MainActivity \
  --es serverUrl "http://192.168.20.26:18080" \
  --es robotCode "robot-001" \
  --es robotToken "rb_p0_xxxxx" \
  --ez autoConnect true'
```

장시간 ADB 자동 테스트에서는 Activity가 foreground에 있어야 실제 카메라 `track.video_1`이 송출된다. 단말이 launcher나 background 상태로 돌아가면 Android camera policy 때문에 RGB 카메라가 열리지 않을 수 있다.

## 현재 상태

현재 Rebuild Phase 11 검증 기준으로 Android Mock은 다음까지 지원한다.

- 관제센터 heartbeat
- active mission polling
- mission 응답 기반 SFU signaling URL/TURN 설정 적용
- mission 응답의 `sfu.signalingUrl`을 그대로 사용
- RGB/Thermal/Audio WebRTC publish
- active mission 중 heartbeat state를 `streaming`으로 유지
- telemetry DataChannel payload는 canonical `descriptors`/`samples`/`values` schema로 송신
- spatial/event/control DataChannel은 negotiation 확인용으로 생성하고 payload는 송신하지 않음
- 위치/가스/배터리 데이터는 WebRTC `channel.telemetry` 경로로만 저장
- React Live UI에서 RGB/Thermal/Audio와 telemetry 수신 확인
- recorder-worker에서 RGB/Thermal/Audio track과 canonical DataChannel open 확인
- recorder-worker에서 RGB/Audio MP4, Thermal MP4 생성과 MinIO 업로드 확인

현재 한계:

- Browser 관제와 recorder-worker 저장의 장시간 동시 수신 안정화는 다음 단계에서 보강한다.
- 실제 Android 단말 검증은 ADB 단말이 연결된 뒤 수행한다.
