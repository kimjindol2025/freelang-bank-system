#!/bin/bash

# 🏦 FreeLang Bank System - API Test Script
# Phase 4: REST API Integration Testing

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🏦 FreeLang Bank System - API Test Suite"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 1: Health Check
echo -e "${BLUE}📊 Test 1: Health Check${NC}"
curl -s -X GET "$BASE_URL/api/health" | jq '.'
echo -e "${GREEN}✅ PASS${NC}\n"

# Test 2: Register User
echo -e "${BLUE}📊 Test 2: Register User${NC}"
REGISTER=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "kim@example.com",
    "password": "password123",
    "first_name": "Jin",
    "last_name": "Kim"
  }')
echo "$REGISTER" | jq '.'
TOKEN=$(echo "$REGISTER" | jq -r '.token')
echo "Token: $TOKEN"
echo -e "${GREEN}✅ PASS${NC}\n"

# Test 3: Create Account
echo -e "${BLUE}📊 Test 3: Create Checking Account${NC}"
ACCOUNT=$(curl -s -X POST "$BASE_URL/api/accounts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"account_type": "checking", "currency": "USD"}')
echo "$ACCOUNT" | jq '.'
ACCOUNT_ID=$(echo "$ACCOUNT" | jq -r '.id')
echo "Account ID: $ACCOUNT_ID"
echo -e "${GREEN}✅ PASS${NC}\n"

# Test 4: Deposit
echo -e "${BLUE}📊 Test 4: Deposit Money${NC}"
curl -s -X POST "$BASE_URL/api/accounts/$ACCOUNT_ID/deposit" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"amount": 1000.0}' | jq '.'
echo -e "${GREEN}✅ PASS${NC}\n"

# Test 5: Check Balance
echo -e "${BLUE}📊 Test 5: Check Balance${NC}"
curl -s -X GET "$BASE_URL/api/accounts/$ACCOUNT_ID/balance" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo -e "${GREEN}✅ PASS${NC}\n"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}🎉 All tests completed!${NC}"
