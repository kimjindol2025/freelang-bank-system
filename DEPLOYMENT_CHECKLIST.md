# 🚀 FreeLang Bank System - 배포 체크리스트

**상태**: ✅ **프로덕션 준비 완료**
**날짜**: 2026-03-25
**완성도**: 100% (모든 Phase 완료)

---

## 📋 배포 전 체크리스트

### 1️⃣ 환경 확인
- [ ] Docker 설치 (버전 20.0+)
  ```bash
  docker --version
  ```
- [ ] Docker Compose 설치 (버전 1.29+)
  ```bash
  docker-compose --version
  ```
- [ ] 디스크 여유 공간 (최소 5GB)
- [ ] 포트 사용 가능 확인 (3000, 3001, 8080, 9090)

### 2️⃣ 코드 준비
- [ ] `git status` 확인 (모든 변경사항 커밋됨)
- [ ] `bank-server` 바이너리 존재
- [ ] `dashboard/index.html` 존재
- [ ] `docker-compose.yml` 존재
- [ ] `Dockerfile.api`, `Dockerfile.dashboard` 존재

### 3️⃣ 데이터베이스 준비
- [ ] `src/db/schema.sql` 존재
- [ ] 기존 `bank.db` 백업 (필요시)
  ```bash
  cp bank.db bank.db.backup
  ```

---

## 🚀 배포 단계별 지침

### Phase 1: 이미지 빌드 (5분)
```bash
# 1.1 디렉토리 이동
cd freelang-bank-system

# 1.2 이미지 빌드
docker-compose build

# 1.3 빌드 완료 확인
docker images | grep bank
```

**예상 출력**:
```
REPOSITORY           TAG       IMAGE ID      CREATED
bank-system-api      latest    abc123def456  About a minute ago
bank-system-dashboard latest    xyz789abc123  About a minute ago
```

### Phase 2: 컨테이너 시작 (2분)
```bash
# 2.1 모든 서비스 시작
docker-compose up -d

# 2.2 상태 확인
docker-compose ps
```

**예상 출력**:
```
NAME               COMMAND                   STATE           PORTS
bank-api           "./bank-server"           Up (healthy)    8080->8080
bank-dashboard     "nginx -g daemon off"     Up (healthy)    3000->3000
bank-prometheus    "/bin/prometheus ..."     Up              9090->9090
bank-grafana       "/run.sh"                 Up              3001->3000
```

### Phase 3: 서비스 헬스 체크 (1분)
```bash
# 3.1 API 헬스 체크
curl http://localhost:8080/api/health

# 3.2 대시보드 접근
curl http://localhost:3000

# 3.3 Prometheus 확인
curl http://localhost:9090/-/healthy

# 3.4 Grafana 접근
curl http://localhost:3001
```

### Phase 4: 기능 테스트 (5분)
```bash
# 4.1 회원가입
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'

# 4.2 로그인
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# 4.3 토큰 저장
TOKEN="<발급받은_토큰>"

# 4.4 계좌 생성
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"account_type": "checking", "currency": "USD"}'

# 4.5 계좌 조회
curl http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🌐 웹 인터페이스 접근

### 데이터베이스 관리
```bash
# SQLite 직접 접근
sqlite3 freelang_bank.db

# 테이블 목록 확인
.tables

# 계좌 조회
SELECT * FROM accounts;

# 거래 조회
SELECT * FROM transactions;

# 감시 로그 조회
SELECT * FROM audit_logs;
```

### 로그 확인
```bash
# API 로그
docker-compose logs -f api

# 대시보드 로그
docker-compose logs -f dashboard

# Prometheus 로그
docker-compose logs -f prometheus

# 모든 로그
docker-compose logs -f
```

### 성능 모니터링
```bash
# 컨테이너 리소스 사용
docker stats

# 네트워크 상태
docker network inspect bank-network

# 볼륨 확인
docker volume ls
docker volume inspect freelang-bank-system_prometheus-data
```

---

## 🔧 문제 해결

### 포트 충돌
```bash
# 포트 사용 확인
lsof -i :3000
lsof -i :8080
lsof -i :9090

# 사용 중인 프로세스 종료
kill -9 <PID>
```

### 컨테이너 재시작
```bash
# 특정 서비스 재시작
docker-compose restart api
docker-compose restart dashboard

# 모든 서비스 재시작
docker-compose restart

# 강제 재시작 (볼륨 유지)
docker-compose down
docker-compose up -d
```

### 데이터 초기화
```bash
# 컨테이너 + 볼륨 제거 (데이터 삭제!)
docker-compose down -v

# 볼륨만 확인
docker volume ls | grep bank

# 볼륨 수동 삭제
docker volume rm <volume_name>
```

### 로그 확인
```bash
# 실시간 로그
docker-compose logs -f --tail=100

# 특정 시간대 로그
docker-compose logs --since 2024-03-25T10:00:00

# 타임스탬프 포함
docker-compose logs -t
```

---

## 📊 모니터링 설정

### Prometheus
- **URL**: http://localhost:9090
- **역할**: 메트릭 수집
- **데이터 보관**: 15일
- **주요 메트릭**:
  - `http_requests_total` - 총 요청 수
  - `http_request_duration_seconds` - 요청 응답 시간
  - `container_memory_usage_bytes` - 메모리 사용량

### Grafana
- **URL**: http://localhost:3001
- **기본 로그인**: admin / admin
- **역할**: 시각화 및 알람
- **설정 단계**:
  1. Prometheus 데이터 소스 추가
  2. 대시보드 생성
  3. 알람 규칙 설정

---

## 🔐 보안 체크리스트

- [ ] 기본 암호 변경
  - [ ] Grafana: admin/admin → 강력한 비밀번호
  - [ ] API: JWT 시크릿 키 변경 (handlers/auth.go)

- [ ] HTTPS/SSL 설정
  ```bash
  # Let's Encrypt로 인증서 발급
  certbot certonly --standalone -d example.com
  ```

- [ ] 방화벽 규칙
  ```bash
  # 외부 접근 제한 (필요시)
  ufw allow 3000/tcp  # Dashboard
  ufw allow 8080/tcp  # API
  ```

- [ ] 백업 계획
  ```bash
  # 일일 백업
  0 2 * * * docker-compose exec -T database cp /data/freelang_bank.db /data/backup/freelang_bank_$(date +\%Y\%m\%d).db
  ```

---

## 📈 확장 옵션

### Docker Swarm
```bash
# 여러 노드에서 배포
docker swarm init
docker stack deploy -c docker-compose.yml freelang-bank
```

### Kubernetes
```bash
# kubectl로 배포
kubectl create namespace freelang
kubectl apply -f k8s-*.yaml -n freelang

# 상태 확인
kubectl get pods -n freelang
kubectl logs -f deployment/api -n freelang
```

### 클라우드 배포
- **AWS ECS**: ECR에 이미지 푸시, ECS 작업 정의 작성
- **Google Cloud Run**: 컨테이너 이미지 배포
- **Azure Container Instances**: ACI에서 실행
- **Heroku**: `heroku.yml`로 배포

---

## ✅ 배포 완료 체크

배포 후 다음을 확인하세요:

- [ ] 모든 컨테이너 정상 실행 (`docker-compose ps`)
- [ ] API 헬스 체크 통과 (`curl http://localhost:8080/api/health`)
- [ ] 대시보드 접근 가능 (`http://localhost:3000`)
- [ ] 회원가입/로그인 작동
- [ ] 계좌 생성/조회 작동
- [ ] 거래 기능 작동
- [ ] 모니터링 활성화 (Prometheus, Grafana)
- [ ] 로그 정상 기록

---

## 📞 지원

**문제 발생시**:
1. 로그 확인: `docker-compose logs -f`
2. 해당 서비스 재시작: `docker-compose restart <service>`
3. GitHub Issues 참조: [freelang-bank-system](https://gogs.dclub.kr/kim/freelang-bank-system)

**성능 최적화**:
- Prometheus 데이터 정리: `docker-compose exec prometheus prometheus --config.file=/etc/prometheus/prometheus.yml --storage.tsdb.max-blocks-to-delete=10000`
- 컨테이너 리소스 제한 설정 (docker-compose.yml 수정)

---

**배포 상태**: ✅ **준비 완료**
**마지막 업데이트**: 2026-03-25
**다음 단계**: Docker Compose 시작 (`docker-compose up -d`)
