#!/usr/bin/env node

import http from 'http';
import url from 'url';
import crypto from 'crypto';

const transactions = new Map();
const batches = new Map();
let batchCount = 0;

function generateId() {
  return crypto.randomUUID();
}

function validAddress(addr) {
  return /^0x[a-fA-F0-9]{40}$/.test(addr);
}

function respond(res, statusCode, data) {
  res.writeHead(statusCode, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(data));
}

function requestHandler(req, res) {
  const parsedUrl = url.parse(req.url, true);
  const pathname = parsedUrl.pathname;

  if (req.method === 'OPTIONS') {
    res.writeHead(200, { 'Access-Control-Allow-Origin': '*' });
    res.end();
    return;
  }

  // GET /
  if (pathname === '/' && req.method === 'GET') {
    respond(res, 200, {
      success: true,
      data: {
        service: 'Banking System (Dev)',
        version: '1.0.0',
        status: 'running'
      },
      timestamp: new Date().toISOString()
    });
    return;
  }

  // GET /api/v1/health
  if (pathname === '/api/v1/health' && req.method === 'GET') {
    respond(res, 200, {
      success: true,
      data: { status: 'healthy' },
      timestamp: new Date().toISOString()
    });
    return;
  }

  // POST /api/v1/webhooks
  if (pathname === '/api/v1/webhooks' && req.method === 'POST') {
    let body = '';
    req.on('data', c => body += c);
    req.on('end', () => {
      try {
        const payload = JSON.parse(body);
        if (!payload.eventId || !payload.data) {
          respond(res, 400, { success: false, error: 'Invalid payload' });
          return;
        }
        if (!validAddress(payload.data.fromAddress) || !validAddress(payload.data.toAddress)) {
          respond(res, 400, { success: false, error: 'Invalid address' });
          return;
        }
        const txId = `txn_${generateId()}`;
        const tx = {
          id: txId,
          eventId: payload.eventId,
          fromAddress: payload.data.fromAddress,
          toAddress: payload.data.toAddress,
          amount: payload.data.amount,
          currency: payload.data.currency,
          status: 'received',
          createdAt: new Date().toISOString()
        };
        transactions.set(txId, tx);
        respond(res, 201, {
          success: true,
          data: tx,
          timestamp: new Date().toISOString()
        });
      } catch (err) {
        respond(res, 500, { success: false, error: 'Server error' });
      }
    });
    return;
  }

  // GET /api/v1/webhooks
  if (pathname === '/api/v1/webhooks' && req.method === 'GET') {
    const txs = Array.from(transactions.values());
    respond(res, 200, {
      success: true,
      data: { transactions: txs, count: txs.length },
      timestamp: new Date().toISOString()
    });
    return;
  }

  // POST /api/v1/settlement/batches
  if (pathname === '/api/v1/settlement/batches' && req.method === 'POST') {
    let body = '';
    req.on('data', c => body += c);
    req.on('end', () => {
      try {
        const payload = JSON.parse(body);
        if (!payload.transactionIds) {
          respond(res, 400, { success: false, error: 'Invalid request' });
          return;
        }
        let total = 0;
        payload.transactionIds.forEach(id => {
          const tx = transactions.get(id);
          if (tx) total += parseFloat(tx.amount);
        });
        batchCount++;
        const batch = {
          id: `batch_${generateId()}`,
          batchNumber: batchCount,
          status: 'pending',
          transactionCount: payload.transactionIds.length,
          totalAmount: total.toString(),
          createdAt: new Date().toISOString()
        };
        batches.set(batch.id, batch);
        respond(res, 201, {
          success: true,
          data: batch,
          timestamp: new Date().toISOString()
        });
      } catch (err) {
        respond(res, 500, { success: false, error: 'Server error' });
      }
    });
    return;
  }

  // GET /api/v1/settlement/batches
  if (pathname === '/api/v1/settlement/batches' && req.method === 'GET') {
    const batchList = Array.from(batches.values());
    respond(res, 200, {
      success: true,
      data: { batches: batchList, count: batchList.length },
      timestamp: new Date().toISOString()
    });
    return;
  }

  // 404
  respond(res, 404, { success: false, error: 'Not found' });
}

const server = http.createServer(requestHandler);
const PORT = 3000;

server.listen(PORT, () => {
  console.log('');
  console.log('🏦 Banking System - Development Server');
  console.log('=====================================');
  console.log(`✅ Running on http://localhost:${PORT}`);
  console.log('');
  console.log('🔗 Available Endpoints:');
  console.log('   GET  /                          - API info');
  console.log('   GET  /api/v1/health             - Health check');
  console.log('   POST /api/v1/webhooks           - Submit payment');
  console.log('   GET  /api/v1/webhooks           - List transactions');
  console.log('   POST /api/v1/settlement/batches - Create batch');
  console.log('   GET  /api/v1/settlement/batches - List batches');
  console.log('');
  console.log('📝 Quick test:');
  console.log(`   curl http://localhost:${PORT}/api/v1/health`);
  console.log('');
});

process.on('SIGINT', () => {
  console.log('\n✅ Server stopped');
  process.exit(0);
});
