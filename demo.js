#!/usr/bin/env node

// 🏦 FreeLang Bank System Demo
// FreeLang 코드의 동작을 JavaScript로 시뮬레이션

console.log("🏦 FreeLang Bank System - Demo");
console.log("=".repeat(50));
console.log("");

// ========================================
// Account Type Definition
// ========================================

class Account {
  constructor(id, name, type, balance = 0.0) {
    this.id = id;
    this.name = name;
    this.balance = balance;
    this.type = type;
    this.rate = this.getRate(type);
  }

  getRate(type) {
    switch(type) {
      case "Savings": return 2.0;
      case "MoneyMarket": return 3.0;
      case "CD": return 5.0;
      default: return 0.0;
    }
  }

  deposit(amount) {
    if (amount <= 0) {
      console.log("❌ 입금액은 0보다 커야 합니다");
      return this;
    }
    return new Account(this.id, this.name, this.type, this.balance + amount);
  }

  withdraw(amount) {
    if (amount <= 0) {
      console.log("❌ 출금액은 0보다 커야 합니다");
      return this;
    }
    if (this.balance < amount) {
      console.log("❌ 잔액 부족");
      return this;
    }
    return new Account(this.id, this.name, this.type, this.balance - amount);
  }

  applyDailyInterest() {
    if (this.rate === 0) return this;
    const interest = this.balance * (this.rate / 100.0 / 365.0);
    return new Account(this.id, this.name, this.type, this.balance + interest);
  }

  applyMonthlyInterest() {
    if (this.rate === 0) return this;
    const interest = this.balance * (this.rate / 100.0 / 12.0);
    return new Account(this.id, this.name, this.type, this.balance + interest);
  }

  transfer(toAccount, amount) {
    const fee = amount > 1000 ? amount * 0.005 : 1.0;
    const fromNew = this.withdraw(amount + fee);
    const toNew = toAccount.deposit(amount);
    return { from: fromNew, to: toNew, fee: fee };
  }

  print() {
    console.log("━".repeat(50));
    console.log(`📋 계좌: ${this.id}`);
    console.log(`👤 예금주: ${this.name}`);
    console.log(`💰 잔액: $${this.balance.toFixed(2)}`);
    console.log(`📊 타입: ${this.type}`);
    console.log(`📈 이율: ${this.rate}% APY`);
    console.log("━".repeat(50));
  }
}

// ========================================
// Demo: Bank Operations
// ========================================

console.log("1️⃣ 계좌 생성");
console.log("-".repeat(50));

let acc1 = new Account("ACC001", "김진돌", "Checking");
let acc2 = new Account("ACC002", "이순신", "Savings");

console.log("✅ Checking 계좌 생성: " + acc1.id);
console.log("✅ Savings 계좌 생성: " + acc2.id);

console.log("");
console.log("2️⃣ 입금");
console.log("-".repeat(50));

acc1 = acc1.deposit(1500.0);
console.log("✅ 김진돌 입금: $1,500");

acc2 = acc2.deposit(5000.0);
console.log("✅ 이순신 입금: $5,000");

acc1.print();
console.log("");
acc2.print();

console.log("");
console.log("3️⃣ 출금");
console.log("-".repeat(50));

acc1 = acc1.withdraw(300.0);
console.log("✅ 김진돌 출금: $300");
console.log(`   잔액: $${acc1.balance.toFixed(2)}`);

console.log("");
console.log("4️⃣ 이자 계산");
console.log("-".repeat(50));

const dailyInterest = acc2.balance * (acc2.rate / 100.0 / 365.0);
console.log(`이순신의 일일 이자: $${dailyInterest.toFixed(6)}`);

acc2 = acc2.applyDailyInterest();
console.log(`✅ 이자 적용 후: $${acc2.balance.toFixed(2)}`);

const monthlyInterest = acc2.balance * (acc2.rate / 100.0 / 12.0);
console.log(`월간 이자: $${monthlyInterest.toFixed(2)}`);

const annualInterest = acc2.balance * (acc2.rate / 100.0);
console.log(`연간 이자: $${annualInterest.toFixed(2)}`);

console.log("");
console.log("5️⃣ 계좌이체");
console.log("-".repeat(50));

const transferAmount = 200.0;
const result = acc1.transfer(acc2, transferAmount);
acc1 = result.from;
acc2 = result.to;

console.log(`✅ 김진돌 → 이순신 이체`);
console.log(`   이체액: $${transferAmount.toFixed(2)}`);
console.log(`   수수료: $${result.fee.toFixed(2)}`);
console.log(`   김진돌 잔액: $${acc1.balance.toFixed(2)}`);
console.log(`   이순신 잔액: $${acc2.balance.toFixed(2)}`);

console.log("");
console.log("6️⃣ 사기 탐지 테스트");
console.log("-".repeat(50));

function detectLargeTransaction(amount) {
  if (amount > 100_000) return "🚨 Critical";
  if (amount > 50_000) return "🔴 High";
  if (amount > 10_000) return "🟡 Medium";
  return "✅ Low";
}

console.log(`거래: $10,000 → ${detectLargeTransaction(10_000)}`);
console.log(`거래: $50,000 → ${detectLargeTransaction(50_000)}`);
console.log(`거래: $150,000 → ${detectLargeTransaction(150_000)}`);

console.log("");
console.log("7️⃣ 최종 잔액");
console.log("-".repeat(50));

acc1.print();
console.log("");
acc2.print();

const total = acc1.balance + acc2.balance;
console.log("");
console.log(`💰 총 자산: $${total.toFixed(2)}`);

console.log("");
console.log("=" * 50);
console.log("✅ 은행 시스템 데모 완료!");
console.log("");
console.log("📊 통계:");
console.log(`  총 거래: 6건 (입금 2, 출금 1, 이자 2, 이체 1)`);
console.log(`  총 수수료: $1.00`);
console.log(`  총 이자 수익: $${(dailyInterest + monthlyInterest).toFixed(2)}`);
