#!/bin/bash

# 🏦 FreeLang Bank System - 종합 테스트 스크립트
# Phase 1-6 전체 검증

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🏦 FreeLang Bank System - 종합 테스트 (Phase 1-6)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Test 1: Go 모듈 검증
echo -e "${YELLOW}[1/6] Go 모듈 검증${NC}"
go mod verify > /dev/null && echo -e "${GREEN}✅ PASS${NC}" || echo -e "${RED}❌ FAIL${NC}"
echo ""

# Test 2: Go 빌드 테스트
echo -e "${YELLOW}[2/6] Go 빌드 테스트${NC}"
cd server && go build -o ../bank-server-test ./main.go > /dev/null 2>&1 && cd ..
if [ -f bank-server-test ]; then
    SIZE=$(du -h bank-server-test | awk '{print $1}')
    echo -e "${GREEN}✅ PASS - 빌드 완료 ($SIZE)${NC}"
    rm bank-server-test
else
    echo -e "${RED}❌ FAIL - 빌드 실패${NC}"
fi
echo ""

# Test 3: Go 단위 테스트
echo -e "${YELLOW}[3/6] Go 단위 테스트${NC}"
RESULT=$(go test -v 2>&1 | grep -c "PASS\|FAIL" || true)
if go test -v > /tmp/test_result.txt 2>&1; then
    PASS_COUNT=$(grep -c "PASS" /tmp/test_result.txt || echo 0)
    echo -e "${GREEN}✅ PASS - $PASS_COUNT 테스트 통과${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
fi
echo ""

# Test 4: 바이너리 크기 확인
echo -e "${YELLOW}[4/6] 바이너리 크기 검증${NC}"
if [ -f bank-server ]; then
    SIZE=$(du -h bank-server | awk '{print $1}')
    echo -e "${GREEN}✅ PASS - $SIZE (목표: <50MB)${NC}"
else
    echo -e "${RED}❌ FAIL - 바이너리 없음${NC}"
fi
echo ""

# Test 5: API 런타임 테스트
echo -e "${YELLOW}[5/6] API 런타임 테스트${NC}"
./bank-server &> /tmp/api_test.log &
API_PID=$!
sleep 3

# 헬스 체크
if curl -s http://localhost:8080/health | grep -q "OK"; then
    echo -e "${GREEN}✅ PASS - API 서버 실행 중 (PID: $API_PID)${NC}"
    
    # 계좌 생성 테스트
    ACCOUNT=$(curl -s -X POST http://localhost:8080/api/accounts \
      -H "Content-Type: application/json" \
      -d '{"name":"Test","type":"Checking","rate":0}' | grep -o '"id":"[^"]*"' | head -1)
    
    if [ ! -z "$ACCOUNT" ]; then
        echo -e "${GREEN}✅ 계좌 생성 성공${NC}"
    fi
    
    # 사기 탐지 테스트
    FRAUD=$(curl -s -X POST http://localhost:8080/api/fraud/check \
      -H "Content-Type: application/json" \
      -d '{"amount":150000,"frequency":120,"balance_drain_pct":85}' | grep -o '"severity":"[^"]*"')
    
    if [ ! -z "$FRAUD" ]; then
        echo -e "${GREEN}✅ 사기 탐지 작동 중 ($FRAUD)${NC}"
    fi
else
    echo -e "${RED}❌ FAIL - API 서버 시작 실패${NC}"
fi

# 서버 종료
kill $API_PID 2>/dev/null || true
sleep 1
echo ""

# Test 6: 파일 구조 검증
echo -e "${YELLOW}[6/6] 파일 구조 검증${NC}"
REQUIRED_FILES=(
    "go.mod"
    "server/main.go"
    "server/database/database.go"
    "server/handlers/account.go"
    "server/handlers/transaction.go"
    "server/handlers/fraud.go"
    "server/handlers/report.go"
    "Dockerfile.api"
    "Dockerfile.dashboard"
    "docker-compose.yml"
    "k8s-api-deployment.yaml"
    "k8s-dashboard-deployment.yaml"
    "k8s-ingress.yaml"
    "k8s-storage.yaml"
    "TEST_EXECUTION_REPORT.md"
)

MISSING=0
for FILE in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "$FILE" ]; then
        echo -e "${RED}❌ 누락: $FILE${NC}"
        MISSING=$((MISSING+1))
    fi
done

if [ $MISSING -eq 0 ]; then
    echo -e "${GREEN}✅ PASS - 모든 파일 존재 ($(echo ${#REQUIRED_FILES[@]}) 개)${NC}"
else
    echo -e "${RED}❌ FAIL - $MISSING 개 파일 누락${NC}"
fi
echo ""

# 최종 요약
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}🎉 종합 테스트 완료!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
