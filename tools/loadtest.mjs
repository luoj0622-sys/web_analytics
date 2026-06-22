#!/usr/bin/env node

const target = process.env.WA_LOADTEST_URL ?? 'http://localhost:8080/collect';
const events = Number(process.env.WA_LOADTEST_EVENTS ?? '1000');
const concurrency = Number(process.env.WA_LOADTEST_CONCURRENCY ?? '50');

let sent = 0;
let failed = 0;

async function sendOne(index) {
  const payload = {
    site_id: process.env.WA_LOADTEST_SITE_ID ?? 'site_loadtest',
    public_key: process.env.WA_LOADTEST_PUBLIC_KEY ?? 'pk_loadtest',
    type: index % 10 === 0 ? 'custom' : 'page_view',
    name: index % 10 === 0 ? 'loadtest_event' : '',
    occurred_at: new Date().toISOString(),
    visitor: {
      id: `visitor_${index % 10000}`,
      session_id: `session_${index % 20000}`
    },
    page: {
      url: `https://example.com/page-${index % 100}`,
      path: `/page-${index % 100}`,
      referrer: 'https://referrer.example'
    },
    campaign: {
      source: 'loadtest',
      medium: 'script',
      campaign: 'capacity'
    },
    device: {
      type: 'desktop',
      browser: 'loadtest',
      os: 'loadtest'
    }
  };

  const response = await fetch(target, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  if (!response.ok && response.status !== 204) {
    failed++;
  }
  sent++;
}

async function worker(offset) {
  for (let i = offset; i < events; i += concurrency) {
    await sendOne(i);
  }
}

const started = Date.now();
await Promise.all(Array.from({ length: concurrency }, (_, index) => worker(index)));
const durationSeconds = (Date.now() - started) / 1000;

console.log(JSON.stringify({
  target,
  events,
  concurrency,
  sent,
  failed,
  duration_seconds: durationSeconds,
  events_per_second: sent / durationSeconds,
  stable_target: '30 million events/day',
  stretch_target: '50 million events/day'
}, null, 2));
