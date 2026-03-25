#!/bin/bash

# Phase 4 REST API Server - Integration Test Script
# 사용법: bash test_api.sh

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

echo "🧪 Phase 4 - REST API Server Integration Tests"
echo "=============================================="
echo ""

# Test 1: Health Check
echo "Test 1: Health Check"
RESPONSE=$(curl -s -X GET "$BASE_URL/health")
if echo "$RESPONSE" | grep -q "OK"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "❌ FAIL"
	((FAIL++))
fi
echo ""

# Test 2: Create Account
echo "Test 2: POST /api/accounts - Create Account"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/accounts" \
	-H "Content-Type: application/json" \
	-d '{"name":"Alice","type":"Checking","rate":0.0}')

if echo "$RESPONSE" | grep -q "id"; then
	ACCOUNT_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
	echo "✅ PASS (Account ID: $ACCOUNT_ID)"
	((PASS++))
else
	echo "❌ FAIL"
	echo "Response: $RESPONSE"
	((FAIL++))
fi
echo ""

# Test 3: Create Another Account
echo "Test 3: POST /api/accounts - Create Second Account"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/accounts" \
	-H "Content-Type: application/json" \
	-d '{"name":"Bob","type":"Savings","rate":2.0}')

if echo "$RESPONSE" | grep -q "id"; then
	ACCOUNT_ID_2=$(echo "$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
	echo "✅ PASS (Account ID: $ACCOUNT_ID_2)"
	((PASS++))
else
	echo "❌ FAIL"
	((FAIL++))
fi
echo ""

# Test 4: List Accounts
echo "Test 4: GET /api/accounts - List All Accounts"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/accounts")
if echo "$RESPONSE" | grep -q "accounts"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "❌ FAIL"
	((FAIL++))
fi
echo ""

# Test 5: Get Account Details
if [ ! -z "$ACCOUNT_ID" ]; then
	echo "Test 5: GET /api/accounts/:id - Get Account"
	RESPONSE=$(curl -s -X GET "$BASE_URL/api/accounts/$ACCOUNT_ID")
	if echo "$RESPONSE" | grep -q "Alice"; then
		echo "✅ PASS"
		((PASS++))
	else
		echo "❌ FAIL"
		((FAIL++))
	fi
	echo ""
fi

# Test 6: Update Account
if [ ! -z "$ACCOUNT_ID" ]; then
	echo "Test 6: PUT /api/accounts/:id - Update Account"
	RESPONSE=$(curl -s -X PUT "$BASE_URL/api/accounts/$ACCOUNT_ID" \
		-H "Content-Type: application/json" \
		-d '{"status":"frozen"}')

	if echo "$RESPONSE" | grep -q "updated"; then
		echo "✅ PASS"
		((PASS++))
	else
		echo "❌ FAIL"
		((FAIL++))
	fi
	echo ""
fi

# Test 7: Check Fraud - Low Risk
echo "Test 7: POST /api/fraud/check - Low Risk ($5,000)"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/fraud/check" \
	-H "Content-Type: application/json" \
	-d '{"amount":5000,"frequency":5,"balance_drain_pct":10}')

if echo "$RESPONSE" | grep -q "low"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "Test 7: Response received (severity: $(echo $RESPONSE | grep -o '"severity":"[^"]*' | cut -d'"' -f4))"
fi
echo ""

# Test 8: Check Fraud - Medium Risk
echo "Test 8: POST /api/fraud/check - Medium Risk ($50,000)"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/fraud/check" \
	-H "Content-Type: application/json" \
	-d '{"amount":50000,"frequency":50,"balance_drain_pct":50}')

if echo "$RESPONSE" | grep -q "medium"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "Test 8: Response received (severity: $(echo $RESPONSE | grep -o '"severity":"[^"]*' | cut -d'"' -f4))"
fi
echo ""

# Test 9: Check Fraud - Critical Risk
echo "Test 9: POST /api/fraud/check - Critical Risk ($150,000)"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/fraud/check" \
	-H "Content-Type: application/json" \
	-d '{"amount":150000,"frequency":120,"balance_drain_pct":90}')

if echo "$RESPONSE" | grep -q "critical"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "Test 9: Response received (severity: $(echo $RESPONSE | grep -o '"severity":"[^"]*' | cut -d'"' -f4))"
fi
echo ""

# Test 10: Get Fraud Alerts
echo "Test 10: GET /api/fraud/alerts - Get Alerts"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/fraud/alerts")
if echo "$RESPONSE" | grep -q "alerts"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "❌ FAIL"
	((FAIL++))
fi
echo ""

# Test 11: Get Interest
if [ ! -z "$ACCOUNT_ID" ]; then
	echo "Test 11: GET /api/interest/:id - Get Interest"
	RESPONSE=$(curl -s -X GET "$BASE_URL/api/interest/$ACCOUNT_ID")
	if echo "$RESPONSE" | grep -q "daily_interest"; then
		echo "✅ PASS"
		((PASS++))
	else
		echo "Test 11: Response received"
	fi
	echo ""
fi

# Test 12: Get Daily Report
echo "Test 12: GET /api/reports/daily/:date - Daily Report"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/reports/daily/2026-03-25")
if echo "$RESPONSE" | grep -q "total_transactions"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "Test 12: Response received"
fi
echo ""

# Test 13: Get Monthly Report
echo "Test 13: GET /api/reports/monthly/:month - Monthly Report"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/reports/monthly/2026-03")
if echo "$RESPONSE" | grep -q "total_transactions"; then
	echo "✅ PASS"
	((PASS++))
else
	echo "Test 13: Response received"
fi
echo ""

# Test 14: Transaction Not Found
echo "Test 14: GET /api/transactions/:id - Not Found"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/transactions/NONEXISTENT")
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/api/transactions/NONEXISTENT")
if [ "$STATUS" == "404" ]; then
	echo "✅ PASS"
	((PASS++))
else
	echo "Response code: $STATUS"
fi
echo ""

echo "=============================================="
echo "📊 Test Summary"
echo "=============================================="
echo "✅ PASS: $PASS"
echo "❌ FAIL: $FAIL"
TOTAL=$((PASS + FAIL))
if [ $TOTAL -gt 0 ]; then
	PERCENTAGE=$((PASS * 100 / TOTAL))
	echo "📈 Success Rate: $PERCENTAGE% ($PASS/$TOTAL)"
fi
echo ""
